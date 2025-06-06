package infrastructure

import (
	"thanhnt208/healthcheck-service/config"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	brokers []string
	writer  *kafka.Writer
	cfg     *config.Config
}

func NewKafka(cfg *config.Config) (*Kafka, error) {
	return &Kafka{
		brokers: cfg.KafkaBrokers,
		cfg:     cfg,
	}, nil
}

func (k *Kafka) ConnectProducer() (*kafka.Writer, error) {
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.brokers...),
		Topic:        k.cfg.KafkaTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}

	return k.writer, nil
}

func (k *Kafka) Close() error {
	if k.writer != nil {
		return k.writer.Close()
	}
	return nil
}
