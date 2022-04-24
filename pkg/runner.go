package pin

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type runner struct {
	ctx               context.Context
	cli               *client.Client
	containerResponse container.ContainerCreateCreatedBody
	currentJob        Job
	workDir           string
	infoLog           *log.Logger
}

func (r *runner) run(workflow Workflow) error {
	r.infoLog = log.New(os.Stdout, "INFO \t", log.Ldate|log.Ltime)
	r.ctx = context.Background()

	cli, err := client.NewClientWithOpts()

	if err != nil {
		return err
	}

	r.cli = cli

	for _, job := range workflow {
		r.currentJob = job
		r.workDir = job.WorkDir

		if err := r.jobRunner(); err != nil {
			return err
		}
	}

	return nil
}

func (r *runner) jobRunner() error {
	isImageAvailable, err := r.checkTheImageAvailable()

	if err != nil {
		return err
	}

	if !isImageAvailable {
		if err := r.pullImage(); err != nil {
			return err
		}
	}

	r.infoLog.Println("Start creating container")

	resp, err := r.cli.ContainerCreate(r.ctx, &container.Config{
		Image: r.currentJob.Image,
		Tty:   true,
	}, nil, nil, nil, r.currentJob.Name)

	if err != nil {
		return err
	}

	r.containerResponse = resp

	if err := r.copyToContainer(); err != nil {
		return err
	}

	r.infoLog.Println("Starting the container")

	if err := r.cli.ContainerStart(r.ctx, r.containerResponse.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if err := r.prepareAndRunShellCommandScript(); err != nil {
		return err
	}

	if err := r.stopCurrentContainer(); err != nil {
		return err
	}

	if err := r.removeCurrentContainer(); err != nil {
		return err
	}

	r.infoLog.Println("Job ended")

	return nil
}

func (r *runner) prepareAndRunShellCommandScript() error {
	shellFileContains := ""

	for _, cmd := range r.currentJob.Script {
		shellFileContains += cmd + "\n"
	}

	if _, err := os.Stat(".pin"); os.IsNotExist(err) {
		err = os.Mkdir(".pin", 0644)

		if err != nil {
			return err
		}
	}

	err := os.WriteFile(".pin/shell_command.sh", []byte(shellFileContains), 0644)

	if err != nil {
		return err
	}

	err = r.sendShellCommandFile()

	if err != nil {
		return err
	}

	if err := r.commandRunner("chmod +x ./shell_command.sh"); err != nil {
		return err
	}

	if err := r.commandRunner("sh ./shell_command.sh"); err != nil {
		return err
	}

	return nil
}

func (r *runner) commandRunner(command string) error {
	args := strings.Split(command, " ")

	r.infoLog.Printf("Execute command: %s", command)

	exec, err := r.cli.ContainerExecCreate(r.ctx, r.containerResponse.ID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          args,
		WorkingDir:   r.workDir,
	})

	if err != nil {
		return err
	}

	res, err := r.cli.ContainerExecAttach(r.ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}

	io.Copy(os.Stdout, res.Reader)

	status, err := r.cli.ContainerExecInspect(r.ctx, exec.ID)
	if err != nil {
		return err
	}

	if status.ExitCode != 0 {
		r.infoLog.Println("Command execution failed")

		r.cli.ContainerKill(r.ctx, r.containerResponse.ID, "KILL")

		// TODO: Print detail exit 1 logs
		out, err := r.cli.ContainerLogs(r.ctx, r.containerResponse.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			return err
		}

		io.Copy(os.Stdout, out)

		if err := r.stopCurrentContainer(); err != nil {
			return err
		}

		if err := r.removeCurrentContainer(); err != nil {
			return err
		}

		return errors.New("command execution failed")
	}

	r.infoLog.Println("Command execution successful")

	return nil
}

func (r runner) checkTheImageAvailable() (bool, error) {
	images, err := r.cli.ImageList(r.ctx, types.ImageListOptions{})

	if err != nil {
		return false, err
	}

	for _, v := range images {
		if r.currentJob.Image == v.RepoTags[0] {
			r.infoLog.Println("Image is available")
			return true, nil
		}
	}

	return false, nil
}

func (r runner) pullImage() error {
	r.infoLog.Printf("Image pulling: %s", r.currentJob.Image)

	reader, err := r.cli.ImagePull(r.ctx, r.currentJob.Image, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

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

func (r runner) sendShellCommandFile() error {
	var buf bytes.Buffer

	data, err := os.ReadFile(".pin/shell_command.sh")

	if err != nil {
		return err
	}

	tw := tar.NewWriter(&buf)
	defer tw.Close()

	err = tw.WriteHeader(&tar.Header{
		Name: "shell_command.sh",
		Mode: 0777,
		Size: int64(len(data)),
	})

	if err != nil {
		return err
	}

	_, err = tw.Write(data)

	if err != nil {
		return err
	}

	err = r.cli.CopyToContainer(r.ctx, r.containerResponse.ID, r.workDir, &buf, types.CopyToContainerOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (r runner) stopCurrentContainer() error {
	r.infoLog.Println("Container stopping")

	if err := r.cli.ContainerStop(r.ctx, r.containerResponse.ID, nil); err != nil {
		return err
	}

	r.infoLog.Println("Container stopped")

	return nil
}

func (r runner) removeCurrentContainer() error {
	r.infoLog.Println("Container removing")

	if err := r.cli.ContainerRemove(r.ctx, r.containerResponse.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	r.infoLog.Println("Container removed")

	return nil
}