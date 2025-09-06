package runner

import (
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/muhammedikinci/pin/internal/interfaces"
)

type Job struct {
	Name             string
	Image            string
	Dockerfile       string
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
	Env              []string
	ArtifactPath     string
	Condition        string
}

type Port struct {
	Out    string // Host port
	In     string // Container port
	HostIP string // Host IP (optional, defaults to 0.0.0.0)
}
