package pin

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
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

	if err := r.copyToContainer(); err != nil {
		return err
	}

	color.Set(color.FgGreen)
	r.infoLog.Println("Starting the container")
	color.Unset()

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

	color.Set(color.FgGreen)
	r.infoLog.Println("Job ended")
	color.Unset()

	return nil
}

func (r *runner) commandScriptExecutor(userCommandLines string) error {
	shellFileContains := "#!/bin/sh\nexec > /shell_command_output.log 2>&1\n" + userCommandLines

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

	if err := r.commandRunner("chmod +x /home/shell_command.sh", ""); err != nil {
		return err
	}

	if err := r.commandRunner("sh /home/shell_command.sh", userCommandLines); err != nil {
		return err
	}

	// not neccessary to handle any error
	os.Remove(".pin/shell_command.sh")

	return nil
}

func (r *runner) commandRunner(command string, name string) error {
	args := strings.Split(command, " ")

	if name != "" && r.currentJob.SoloExecution {
		r.infoLog.Printf("Execute command: %s", name)
	} else if !r.currentJob.SoloExecution {
		r.infoLog.Println("soloExecution disabled, shell command started!")
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

	res, err := r.cli.ContainerExecAttach(r.ctx, exec.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return err
	}

	io.Copy(os.Stdout, res.Reader)

	status, err := r.cli.ContainerExecInspect(r.ctx, exec.ID)
	if err != nil {
		return err
	}

	if status.ExitCode != 0 {
		color.Set(color.FgRed)
		r.infoLog.Printf("Command execution failed")

		r.infoLog.Println("=======================")
		r.infoLog.Println("Command Log:")

		if reader, _, err := r.cli.CopyFromContainer(r.ctx, r.containerResponse.ID, "/shell_command_output.log"); err == nil {
			io.Copy(os.Stdout, reader)
		}
		r.infoLog.Println("=======================")
		color.Unset()

		r.cli.ContainerKill(r.ctx, r.containerResponse.ID, "KILL")

		if err := r.stopCurrentContainer(); err != nil {
			return err
		}

		if err := r.removeCurrentContainer(); err != nil {
			return err
		}

		return errors.New("command execution failed")
	}

	r.infoLog.Println("Command execution successful")

	if reader, _, err := r.cli.CopyFromContainer(r.ctx, r.containerResponse.ID, "/shell_command_output.log"); err == nil {
		var buf bytes.Buffer

		io.Copy(&buf, reader)

		if buf.Len() != 0 {
			color.Set(color.FgGreen)
			r.infoLog.Println("=======================")
			r.infoLog.Println("Command Log:")
			io.Copy(os.Stdout, &buf)
			r.infoLog.Println("=======================")
			color.Unset()
		}
	}

	return nil
}
