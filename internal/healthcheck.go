package internal

import (
	"context"
	"fmt"
	"thanhnt208/healthcheck-service/external/client"
)

func IsContainerRunning(ctx context.Context, containerName string) (bool, error) {
	cli, err := client.NewDockerClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Docker client: %w", err)
	}

	status, err := cli.InspectContainer(ctx, containerName)
	if err != nil {
		return false, fmt.Errorf("failed to inspect container %s: %w", containerName, err)
	}
	return status, nil
}
