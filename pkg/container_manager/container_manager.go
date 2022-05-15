package container_manager

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/pkg/interfaces"
)

type containerManager struct {
	cli interfaces.Client
	log interfaces.Log
}

func NewContainerManager(cli interfaces.Client, log interfaces.Log) containerManager {
	return containerManager{
		cli: cli,
		log: log,
	}
}

func (cm containerManager) StartContainer(ctx context.Context, jobName string, image string, ports map[string]string) (container.ContainerCreateCreatedBody, error) {
	color.Set(color.FgGreen)
	cm.log.Println("Start creating container")
	color.Unset()

	containerName := jobName + "_" + strconv.Itoa(int(time.Now().UnixMilli()))

	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for out, in := range ports {
		inPort, _ := nat.NewPort("tcp", in)

		if _, ok := portBindings[inPort]; ok {
			portBindings[inPort] = append(portBindings[inPort], nat.PortBinding{HostIP: "0.0.0.0", HostPort: out})
		} else {
			portBindings[inPort] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: out}}
		}

		exposedPorts[inPort] = struct{}{}
	}

	hostConfig := &container.HostConfig{PortBindings: portBindings}

	resp, err := cm.cli.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Tty:          true,
		ExposedPorts: exposedPorts,
	}, hostConfig, nil, nil, containerName)

	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}

	return resp, nil
}

func (cm containerManager) StopContainer(ctx context.Context, containerID string) error {
	color.Set(color.FgBlue)
	cm.log.Println("Container stopping")

	if err := cm.cli.ContainerStop(ctx, containerID, nil); err != nil {
		return err
	}

	cm.log.Println("Container stopped")
	color.Unset()

	return nil
}

func (cm containerManager) RemoveContainer(ctx context.Context, containerID string, forceRemove bool) error {
	color.Set(color.FgBlue)
	cm.log.Println("Container removing")

	if err := cm.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: forceRemove}); err != nil {
		return err
	}

	cm.log.Println("Container removed")
	color.Unset()

	return nil
}

func (cm containerManager) CopyToContainer(ctx context.Context, containerID, workDir string, copyIgnore []string) error {
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	defer tw.Close()

	currentPath, _ := os.Getwd()

	err := filepath.Walk(currentPath, func(path string, info os.FileInfo, err error) error {
		return cm.appender(path, info, err, currentPath, tw, copyIgnore)
	})

	if err != nil {
		return err
	}

	err = cm.cli.CopyToContainer(ctx, containerID, workDir, &buf, types.CopyToContainerOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (cm containerManager) appender(path string, info os.FileInfo, err error, currentPath string, tw *tar.Writer, copyIgnore []string) error {
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return nil
	}

	for _, ignore := range copyIgnore {
		if info.IsDir() && info.Name() == ignore {
			return filepath.SkipDir
		}
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = strings.TrimPrefix(strings.Replace(path, currentPath, "", -1), string(filepath.Separator))
	header.Name = strings.ReplaceAll(header.Name, "\\", "/")

	for _, ignore := range copyIgnore {
		if mathced, err := regexp.MatchString(ignore, header.Name); err != nil || mathced {
			return nil
		}
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
}
