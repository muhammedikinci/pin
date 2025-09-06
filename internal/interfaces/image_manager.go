package interfaces

import "context"

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ImageManager interface {
	CheckTheImageAvailable(ctx context.Context, image string) (bool, error)
	PullImage(ctx context.Context, image string) error
	BuildImageFromDockerfile(ctx context.Context, dockerfilePath string, imageName string) error
}
