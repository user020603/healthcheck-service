package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"thanhnt208/healthcheck-service/config"
	"thanhnt208/healthcheck-service/external/client"
	"thanhnt208/healthcheck-service/infrastructure"
	"thanhnt208/healthcheck-service/pkg/logger"
	"thanhnt208/healthcheck-service/proto/pb"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.LoadConfig()
	if cfg == nil {
		panic("Failed to load configuration")
	}

	logger, err := logger.NewLogger(cfg.LogLevel, cfg.LogFile)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	kafkaInfra, err := infrastructure.NewKafka(cfg)
	if err != nil {
		logger.Error("Failed to create Kafka infrastructure", "error", err)
		panic("Failed to create Kafka infrastructure: " + err.Error())
	}

	producer, err := kafkaInfra.ConnectProducer()
	if err != nil {
		logger.Error("Failed to connect Kafka producer", "error", err)
		panic("Failed to connect Kafka producer: " + err.Error())
	}
	defer func() {
		if err := kafkaInfra.Close(); err != nil {
			logger.Error("Failed to close Kafka writer", "error", err)
		}
	}()

	grpcConn, err := client.StartGrpcClient()
	if err != nil {
		logger.Error("Failed to start gRPC client", "error", err)
		panic("Failed to start gRPC client: " + err.Error())
	}

	healthCheckClient := client.NewHealthCheckClient(grpcConn, logger)

	maxConcurrency, err := strconv.Atoi(cfg.MaxConcurrency)
	if err != nil {
		logger.Error("Failed to parse MaxConcurrency", "error", err)
		panic("Failed to parse MaxConcurrency: " + err.Error())
	}
	delaySeconds, err := strconv.Atoi(cfg.DelaySeconds)
	if err != nil {
		logger.Error("Failed to parse DelaySeconds", "error", err)
		panic("Failed to parse DelaySeconds: " + err.Error())
	}

	semaphore := make(chan struct{}, maxConcurrency)

	dockerClient, err := client.NewDockerClient()
	if err != nil {
		logger.Error("Failed to create Docker client", "error", err)
		panic("Failed to create Docker client: " + err.Error())
	}

	fmt.Printf("Starting health check service with MaxConcurrency: %d, DelaySeconds: %d\n", maxConcurrency, delaySeconds)
	for {
		containers, err := healthCheckClient.GetAllContainers()
		if err != nil {
			logger.Error("Failed to get all containers", "error", err)
			logger.Error("Retrying after delay", "delay", delaySeconds)
			time.Sleep(time.Duration(delaySeconds) * time.Second)
			continue
		}

		logger.Info("Retrieved all containers", "count", len(containers.Containers))

		var wg sync.WaitGroup
		for _, container := range containers.Containers {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(container *pb.ContainerName) {
				defer wg.Done()
				defer func() { <-semaphore }()

				id := int(container.Id)
				containerName := container.ContainerName

				logger.Info("Processing container", "id", id, "name", containerName)
				status, err := dockerClient.InspectContainer(context.Background(), containerName)
				if err != nil {
					logger.Error("Failed to inspect container", "id", id, "name", containerName, "error", err)
					return
				}

				statusText := "stopped"
				if status {
					statusText = "running"
				}
				logger.Info("Container status", "id", id, "name", containerName, "status", statusText)

				payload := map[string]interface{}{
					"id":             id,
					"container_name": containerName,
					"status":         status,
				}

				msg, err := json.Marshal(payload)
				if err != nil {
					logger.Error("Failed to marshal message", "id", id, "name", containerName, "error", err)
					return
				}
				err = producer.WriteMessages(
					context.Background(),
					kafka.Message{
						Key:   []byte(strconv.Itoa(id)),
						Value: msg,
					})
				if err != nil {
					logger.Error("Failed to write message to Kafka", "id", id, "name", containerName, "error", err)
					return
				}

				logger.Info("Message written to Kafka", "id", id, "name", containerName)
			}(container)
		}

		wg.Wait()
		logger.Info("Sleeping for next iteration", "delay", delaySeconds)
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}
}
