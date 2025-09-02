package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func NewDockerManager() (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	return &DockerManager{Cli: cli}, nil
}

func (dm *DockerManager) RefreshContainers() ([]Container, error) {
	containers, err := dm.Cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []Container

	for _, cont := range containers {
		projectName := cont.Labels["com.docker.compose.project"]
		if projectName == "" {
			projectName = "N/A"
		}

		containerName := "N/A"
		if len(cont.Names) > 0 {
			containerName = strings.TrimPrefix(cont.Names[0], "/")
		}

		status := strings.Split(cont.Status, " ")[0]

		c := Container{
			ID:        cont.ID[:12],
			Name:      containerName,
			Status:    status,
			CreatedAt: time.Unix(cont.Created, 0).Format(time.RFC3339),
			Project:   projectName,
			Image:     cont.Image,
			Command:   cont.Command,
			Ports:     cont.Ports,
		}
		result = append(result, c)
	}

	return result, nil
}
