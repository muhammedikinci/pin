package interfaces

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ContainerManager interface {
	StartContainer(ctx context.Context, jobName string, image string, ports map[string]string) (container.ContainerCreateCreatedBody, error)
	StopContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string, forceRemove bool) error
	CopyToContainer(ctx context.Context, containerID, workDir string) error
}
