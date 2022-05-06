package pin

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/fatih/color"
)

func (r *runner) startContainer() error {
	color.Set(color.FgGreen)
	r.infoLog.Println("Start creating container")
	color.Unset()

	resp, err := r.cli.ContainerCreate(r.ctx, &container.Config{
		Image: r.currentJob.Image,
		Tty:   true,
	}, nil, nil, nil, r.currentJob.Name)

	if err != nil {
		return err
	}

	r.containerResponse = resp

	return nil
}

func (r runner) stopCurrentContainer() error {
	color.Set(color.FgBlue)
	r.infoLog.Println("Container stopping")

	if err := r.cli.ContainerStop(r.ctx, r.containerResponse.ID, nil); err != nil {
		return err
	}

	r.infoLog.Println("Container stopped")
	color.Unset()

	return nil
}

func (r runner) removeCurrentContainer() error {
	color.Set(color.FgBlue)
	r.infoLog.Println("Container removing")

	if err := r.cli.ContainerRemove(r.ctx, r.containerResponse.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	r.infoLog.Println("Container removed")
	color.Unset()

	return nil
}

func (r runner) copyToContainer() error {
	if !r.currentJob.CopyFiles {
		return nil
	}

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

	err = r.cli.CopyToContainer(r.ctx, r.containerResponse.ID, r.workDir, &buf, types.CopyToContainerOptions{})

	if err != nil {
		return err
	}

	return nil
}
