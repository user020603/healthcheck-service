package client

import (
	"context"
	"fmt"
	"thanhnt208/healthcheck-service/pkg/logger"
	"thanhnt208/healthcheck-service/proto/pb"
)

type IHealthCheckClient interface {
	GetAllContainers() (*pb.ContainerResponse, error)
}

type healthcheckClient struct {
	client pb.ContainerAdmServiceClient
	logger logger.ILogger
}

func NewHealthCheckClient(client pb.ContainerAdmServiceClient, logger logger.ILogger) IHealthCheckClient {
	return &healthcheckClient{
		client: client,
		logger: logger,
	}
}

func (h *healthcheckClient) GetAllContainers() (*pb.ContainerResponse, error) {
	resp, err := h.client.GetAllContainers(
		context.Background(),
		&pb.EmptyRequest{},
	)

	if err != nil {
		h.logger.Error("Failed to get all containers", "error", err)
		return nil, fmt.Errorf("failed to get all containers: %w", err)
	}

	h.logger.Info("Successfully retrieved all containers", "count", len(resp.Containers))
	return resp, nil
}
