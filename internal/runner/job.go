package runner

import (
	"github.com/docker/docker/api/types/container"
	"github.com/muhammedikinci/pin/internal/container_manager"
	"github.com/muhammedikinci/pin/internal/image_manager"
	"github.com/muhammedikinci/pin/internal/log"
	"github.com/muhammedikinci/pin/internal/shell_commander"
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
	Container        container.CreateResponse
	InfoLog          log.Log
	ImageManager     image_manager.ImageManager
	ContainerManager container_manager.ContainerManager
	ShellCommander   shell_commander.ShellCommander
	Env              []string
	ArtifactPath     string
	Condition        string
	RetryConfig      RetryConfig
}

type RetryConfig struct {
	MaxAttempts int
	DelaySeconds int
	BackoffMultiplier float64
}

type Port struct {
	Out    string // Host port
	In     string // Container port
	HostIP string // Host IP (optional, defaults to 0.0.0.0)
}
