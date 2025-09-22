package test

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
)

// DockerClientMock is a mock interface for the Docker client
type DockerClientMock interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]any, error)
	ContainerInspect(ctx context.Context, containerID string) (any, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (any, error)
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	Ping(ctx context.Context) (any, error)
}

// MockDockerClient implements DockerClientMock for testing
type MockDockerClient struct {
	ContainerListFunc    func(ctx context.Context, options container.ListOptions) ([]any, error)
	ContainerInspectFunc func(ctx context.Context, containerID string) (any, error)
	ContainerStatsFunc   func(ctx context.Context, containerID string, stream bool) (any, error)
	ContainerRestartFunc func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerStartFunc   func(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStopFunc    func(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerPauseFunc   func(ctx context.Context, containerID string) error
	ContainerUnpauseFunc func(ctx context.Context, containerID string) error
	ContainerRemoveFunc  func(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerLogsFunc    func(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	PingFunc             func(ctx context.Context) (any, error)
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]any, error) {
	if m.ContainerListFunc != nil {
		return m.ContainerListFunc(ctx, options)
	}
	return nil, nil
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (any, error) {
	if m.ContainerInspectFunc != nil {
		return m.ContainerInspectFunc(ctx, containerID)
	}
	return nil, nil
}

func (m *MockDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (any, error) {
	if m.ContainerStatsFunc != nil {
		return m.ContainerStatsFunc(ctx, containerID, stream)
	}
	return nil, nil
}

func (m *MockDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.ContainerRestartFunc != nil {
		return m.ContainerRestartFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	if m.ContainerStartFunc != nil {
		return m.ContainerStartFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.ContainerStopFunc != nil {
		return m.ContainerStopFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerPause(ctx context.Context, containerID string) error {
	if m.ContainerPauseFunc != nil {
		return m.ContainerPauseFunc(ctx, containerID)
	}
	return nil
}

func (m *MockDockerClient) ContainerUnpause(ctx context.Context, containerID string) error {
	if m.ContainerUnpauseFunc != nil {
		return m.ContainerUnpauseFunc(ctx, containerID)
	}
	return nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.ContainerRemoveFunc != nil {
		return m.ContainerRemoveFunc(ctx, containerID, options)
	}
	return nil
}

func (m *MockDockerClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	if m.ContainerLogsFunc != nil {
		return m.ContainerLogsFunc(ctx, containerID, options)
	}
	return nil, nil
}

func (m *MockDockerClient) Ping(ctx context.Context) (any, error) {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil, nil
}

func (m *MockDockerClient) Close() error {
	return nil
}

// Test utilities for creating mock responses
func CreateMockContainer(id, name, image, status, state string) map[string]any {
	return map[string]any{
		"ID":      id,
		"Names":   []string{"/" + name},
		"Image":   image,
		"Status":  status,
		"State":   state,
		"Created": time.Now().Unix(),
		"Ports":   []any{},
		"Labels": map[string]string{
			"com.docker.compose.project": "test-project",
		},
	}
}

func CreateMockContainerJSON(id, name, image, status string) map[string]any {
	now := time.Now()
	return map[string]any{
		"ContainerJSONBase": map[string]any{
			"ID":    id,
			"Name":  "/" + name,
			"Image": image,
			"State": map[string]any{
				"Status":    status,
				"StartedAt": now.Format(time.RFC3339Nano),
			},
			"Created": now.Format(time.RFC3339Nano),
		},
		"Config": map[string]any{
			"Image": image,
			"Cmd":   []string{"nginx", "-g", "daemon off;"},
			"Labels": map[string]string{
				"com.docker.compose.project": "test-project",
			},
			"Env": []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		},
		"NetworkSettings": map[string]any{
			"Networks": map[string]any{
				"bridge": map[string]any{
					"IPAddress": "172.17.0.2",
					"Gateway":   "172.17.0.1",
				},
			},
			"Ports": map[string]any{},
		},
		"HostConfig": map[string]any{},
	}
}

func CreateMockStatsResponse() any {
	return map[string]any{
		"Body": io.NopCloser(strings.NewReader(`{
			"cpu_stats": {
				"cpu_usage": {
					"total_usage": 1000000000,
					"percpu_usage": [500000000, 500000000]
				},
				"system_cpu_usage": 20000000000
			},
			"precpu_stats": {
				"cpu_usage": {
					"total_usage": 500000000
				},
				"system_cpu_usage": 10000000000
			},
			"memory_stats": {
				"usage": 100000000,
				"limit": 200000000
			},
			"networks": {
				"eth0": {
					"rx_bytes": 1000,
					"tx_bytes": 2000
				}
			},
			"blkio_stats": {
				"io_service_bytes_recursive": [
					{"op": "Read", "value": 500},
					{"op": "Write", "value": 300}
				]
			}
		}`)),
	}
}
