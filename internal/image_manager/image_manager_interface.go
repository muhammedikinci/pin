package image_manager

import "context"

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_image_manager.go -package mocks
type ImageManager interface {
	CheckTheImageAvailable(ctx context.Context, image string) (bool, error)
	PullImage(ctx context.Context, image string) error
	BuildImageFromDockerfile(ctx context.Context, dockerfilePath string, imageName string) error
}