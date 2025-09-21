package client

import (
	"context"
	"io"

	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_client.go -package mocks
type Client interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error
	CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error)
	ImagePull(ctx context.Context, refStr string, options imagetypes.PullOptions) (io.ReadCloser, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
	ImageList(ctx context.Context, options imagetypes.ListOptions) ([]imagetypes.Summary, error)
	ContainerKill(ctx context.Context, containerID string, signal string) error
}

type dockerClientWrapper struct {
	*dockerclient.Client
}

func NewClient(dockerCli *dockerclient.Client) Client {
	return &dockerClientWrapper{dockerCli}
}

func (w *dockerClientWrapper) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	return w.Client.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (w *dockerClientWrapper) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return w.Client.ContainerStop(ctx, containerID, options)
}

func (w *dockerClientWrapper) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return w.Client.ContainerRemove(ctx, containerID, options)
}

func (w *dockerClientWrapper) CopyToContainer(ctx context.Context, containerID string, dstPath string, content io.Reader, options container.CopyToContainerOptions) error {
	return w.Client.CopyToContainer(ctx, containerID, dstPath, content, options)
}

func (w *dockerClientWrapper) CopyFromContainer(ctx context.Context, containerID string, srcPath string) (io.ReadCloser, container.PathStat, error) {
	return w.Client.CopyFromContainer(ctx, containerID, srcPath)
}

func (w *dockerClientWrapper) ImagePull(ctx context.Context, refStr string, options imagetypes.PullOptions) (io.ReadCloser, error) {
	return w.Client.ImagePull(ctx, refStr, options)
}

func (w *dockerClientWrapper) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	return w.Client.ImageBuild(ctx, buildContext, options)
}

func (w *dockerClientWrapper) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return w.Client.ContainerStart(ctx, containerID, options)
}

func (w *dockerClientWrapper) ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error) {
	return w.Client.ContainerExecCreate(ctx, container, config)
}

func (w *dockerClientWrapper) ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error) {
	return w.Client.ContainerExecAttach(ctx, execID, config)
}

func (w *dockerClientWrapper) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	return w.Client.ContainerExecInspect(ctx, execID)
}

func (w *dockerClientWrapper) ImageList(ctx context.Context, options imagetypes.ListOptions) ([]imagetypes.Summary, error) {
	return w.Client.ImageList(ctx, options)
}

func (w *dockerClientWrapper) ContainerKill(ctx context.Context, containerID string, signal string) error {
	return w.Client.ContainerKill(ctx, containerID, signal)
}