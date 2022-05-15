package pin

import (
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/muhammedikinci/pin/pkg/interfaces"
)

type Job struct {
	Name             string
	Image            string
	Script           []string
	WorkDir          string
	CopyFiles        bool
	SoloExecution    bool
	Port             []Port
	CopyIgnore       []string
	IsParallel       bool
	Previous         *Job
	ErrorChannel     chan error
	Container        container.ContainerCreateCreatedBody
	InfoLog          *log.Logger
	ImageManager     interfaces.ImageManager
	ContainerManager interfaces.ContainerManager
	ShellCommander   interfaces.ShellCommander
}

type Port struct {
	Out string
	In  string
}
