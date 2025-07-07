package interfaces

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ContainerManager interface {
	StartContainer(ctx context.Context, name string, image string, ports map[string]string, env []string) (container.ContainerCreateCreatedBody, error)
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string, forceRemove bool) error
	CopyToContainer(ctx context.Context, containerID string, workDir string, copyIgnore []string) error
	CopyFromContainer(ctx context.Context, containerID string, srcPath string, destPath string) error
}
