package container_manager

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/pkg/interfaces"
)

type ContainerManager struct {
	ctx context.Context
	cli interfaces.Client
	log interfaces.Log
}

func NewContainerManager(ctx context.Context, cli interfaces.Client, log interfaces.Log) ContainerManager {
	return ContainerManager{
		ctx: ctx,
		cli: cli,
		log: log,
	}
}

func (cm ContainerManager) StartContainer(jobName string, image string) (container.ContainerCreateCreatedBody, error) {
	color.Set(color.FgGreen)
	cm.log.Println("Start creating container")
	color.Unset()

	containerName := jobName + "_" + strconv.Itoa(int(time.Now().UnixMilli()))

	resp, err := cm.cli.ContainerCreate(cm.ctx, &container.Config{
		Image: image,
		Tty:   true,
	}, nil, nil, nil, containerName)

	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}

	return resp, nil
}

func (cm ContainerManager) StopContainer(containerID string) error {
	color.Set(color.FgBlue)
	cm.log.Println("Container stopping")

	if err := cm.cli.ContainerStop(cm.ctx, containerID, nil); err != nil {
		return err
	}

	cm.log.Println("Container stopped")
	color.Unset()

	return nil
}

func (cm ContainerManager) RemoveContainer(containerID string) error {
	color.Set(color.FgBlue)
	cm.log.Println("Container removing")

	if err := cm.cli.ContainerRemove(cm.ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	cm.log.Println("Container removed")
	color.Unset()

	return nil
}

func (cm ContainerManager) CopyToContainer(containerID, workDir string) error {
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	defer tw.Close()

	currentPath, _ := os.Getwd()

	// TODO: add dirs, directories does not extract from docker api
	err := filepath.Walk(currentPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(path, currentPath, "", -1), string(filepath.Separator))
		header.Name = strings.ReplaceAll(header.Name, "\\", "/")

		if header.Name[0] == '.' {
			return nil
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = cm.cli.CopyToContainer(cm.ctx, containerID, workDir, &buf, types.CopyToContainerOptions{})

	if err != nil {
		return err
	}

	return nil
}
