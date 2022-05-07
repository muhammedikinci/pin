package interfaces

import "github.com/docker/docker/api/types/container"

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ContainerManager interface {
	StartContainer(jobName string, image string) (container.ContainerCreateCreatedBody, error)
	StopContainer(containerID string) error
	RemoveContainer(containerID string) error
	CopyToContainer(containerID, workDir string) error
}
