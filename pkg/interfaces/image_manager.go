package interfaces

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ImageManager interface {
	CheckTheImageAvailable(image string) (bool, error)
	PullImage(image string) error
}
