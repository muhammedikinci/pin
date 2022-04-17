package pin

import (
	"context"
	"fmt"
	"io"
	"os"
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
}

func (r *runner) run(workflow Workflow) error {
	r.workDir = "/"
	r.ctx = context.Background()

	cli, err := client.NewClientWithOpts()

	if err != nil {
		return err
	}

	r.cli = cli

	for _, job := range workflow {
		r.currentJob = job
		r.workDir = "/"

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

	fmt.Println("Start creating container")

	resp, err := r.cli.ContainerCreate(r.ctx, &container.Config{
		Image: r.currentJob.Image,
		Tty:   true,
	}, nil, nil, nil, "")

	if err != nil {
		return err
	}

	r.containerResponse = resp

	fmt.Println("Starting the container")

	if err := r.cli.ContainerStart(r.ctx, r.containerResponse.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	for _, cmd := range r.currentJob.Script {
		if err := r.commandRunner(cmd); err != nil {
			return err
		}
	}

	if err := r.cli.ContainerStop(r.ctx, r.containerResponse.ID, nil); err != nil {
		return err
	}

	if err := r.cli.ContainerRemove(r.ctx, r.containerResponse.ID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	fmt.Println("Job ended")

	return nil
}

func (r *runner) commandRunner(command string) error {
	args := strings.Split(command, " ")

	if args[0] == "cd" && len(args) == 2 {
		r.workDir = args[1]
		return nil
	}

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

	return nil
}

func (r runner) checkTheImageAvailable() (bool, error) {
	images, err := r.cli.ImageList(r.ctx, types.ImageListOptions{})

	if err != nil {
		return false, err
	}

	for _, v := range images {
		if r.currentJob.Image == v.RepoTags[0] {
			fmt.Println("Image is available")
			return true, nil
		}
	}

	return false, nil
}

func (r runner) pullImage() error {
	fmt.Println("Image pulling: " + r.currentJob.Image)

	reader, err := r.cli.ImagePull(r.ctx, r.currentJob.Image, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

	return nil
}
