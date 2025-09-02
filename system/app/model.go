package app

import (
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerManager struct {
	Cli *client.Client
}

type Container struct {
	ID        string
	Name      string
	Image     string
	Project   string
	Command   string
	CreatedAt string
	Status    string
	Ports     []container.Port
}

type ContainerMsg []Container