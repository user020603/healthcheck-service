package client

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

type IDockerClient interface {
	InspectContainer(ctx context.Context, containerName string) (bool, error)
}

type dockerClient struct {
	client *client.Client
}

func NewDockerClient() (IDockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &dockerClient{client: cli}, nil
}

func (d *dockerClient) InspectContainer(ctx context.Context, containerName string) (bool, error) {
	containerJSON, err := d.client.ContainerInspect(ctx, containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to inspect container %s: %w", containerName, err)
	}

	isRunning := containerJSON.State != nil && containerJSON.State.Running
	return isRunning, nil
}
