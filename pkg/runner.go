package pin

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/pkg/container_manager"
	"github.com/muhammedikinci/pin/pkg/image_manager"
	"github.com/muhammedikinci/pin/pkg/interfaces"
	"github.com/muhammedikinci/pin/pkg/shell_commander"
)

type Runner struct {
	ctx              context.Context
	cli              interfaces.Client
	currentJob       Job
	workDir          string
	infoLog          *log.Logger
	imageManager     interfaces.ImageManager
	containerManager interfaces.ContainerManager
	shellCommander   interfaces.ShellCommander
	container        container.ContainerCreateCreatedBody
}

func (r *Runner) run(workflow Workflow) error {
	r.infoLog = log.New(os.Stdout, "INFO \t", log.Ldate|log.Ltime)
	r.ctx = context.Background()

	cli, err := client.NewClientWithOpts()

	if err != nil {
		return err
	}

	r.cli = cli
	r.imageManager = image_manager.NewImageManager(r.ctx, r.cli, r.infoLog)
	r.containerManager = container_manager.NewContainerManager(r.ctx, r.cli, r.infoLog)

	for _, job := range workflow {
		r.currentJob = job
		r.workDir = job.WorkDir
		r.shellCommander = shell_commander.NewShellCommander()

		if err := r.jobRunner(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) jobRunner() error {
	isImageAvailable, err := r.imageManager.CheckTheImageAvailable(r.currentJob.Image)

	if err != nil {
		return err
	}

	if !isImageAvailable {
		if err := r.imageManager.PullImage(r.currentJob.Image); err != nil {
			return err
		}
	}

	resp, err := r.containerManager.StartContainer(r.currentJob.Name, r.currentJob.Image)

	if err != nil {
		return err
	}

	r.container = resp

	if r.currentJob.CopyFiles {
		if err := r.containerManager.CopyToContainer(resp.ID, r.workDir); err != nil {
			return err
		}
	}

	color.Set(color.FgGreen)
	r.infoLog.Println("Starting the container")
	color.Unset()

	if err := r.cli.ContainerStart(r.ctx, r.container.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if err := r.commandScriptExecutor(); err != nil {
		return err
	}

	if err := r.containerManager.StopContainer(r.container.ID); err != nil {
		return err
	}

	if err := r.containerManager.RemoveContainer(r.container.ID); err != nil {
		return err
	}

	color.Set(color.FgGreen)
	r.infoLog.Println("Job ended")
	color.Unset()

	return nil
}

func (r *Runner) commandScriptExecutor() error {
	cmds := r.shellCommander.PrepareShellCommands(r.currentJob.SoloExecution, r.currentJob.Script)

	for _, cmd := range cmds {
		buf, err := r.shellCommander.ShellToTar(cmd)

		if err != nil {
			return err
		}

		err = r.cli.CopyToContainer(r.ctx, r.container.ID, "/home/", buf, types.CopyToContainerOptions{})

		if err != nil {
			return err
		}

		if err := r.internalExec("chmod +x /home/shell_command.sh"); err != nil {
			return err
		}

		if err := r.commandRunner("sh /home/shell_command.sh", cmd); err != nil {
			return err
		}

		if err := r.internalExec("rm /home/shell_command.sh"); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) commandRunner(command string, name string) error {
	args := strings.Split(command, " ")

	if name != "" && r.currentJob.SoloExecution {
		lines := strings.Split(name, "\n")
		name = strings.Join(lines[2:], "\n")
		r.infoLog.Printf("Execute command: %s", name)
	} else if !r.currentJob.SoloExecution {
		r.infoLog.Println("soloExecution disabled, shell command started!")
	}

	exec, err := r.cli.ContainerExecCreate(r.ctx, r.container.ID, types.ExecConfig{
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

		if reader, _, err := r.cli.CopyFromContainer(r.ctx, r.container.ID, "/shell_command_output.log"); err == nil {
			tr := tar.NewReader(reader)
			tr.Next()
			b, _ := ioutil.ReadAll(tr)
			fmt.Println("\n" + string(b))
		}
		r.infoLog.Println("=======================")
		color.Unset()

		r.cli.ContainerKill(r.ctx, r.container.ID, "KILL")

		if err := r.containerManager.StopContainer(r.container.ID); err != nil {
			return err
		}

		if err := r.containerManager.RemoveContainer(r.container.ID); err != nil {
			return err
		}

		return errors.New("command execution failed")
	}

	r.infoLog.Println("Command execution successful")

	if reader, _, err := r.cli.CopyFromContainer(r.ctx, r.container.ID, "/shell_command_output.log"); err == nil {
		tr := tar.NewReader(reader)
		tr.Next()
		b, _ := ioutil.ReadAll(tr)

		if len(b) != 0 {
			color.Set(color.FgGreen)
			r.infoLog.Println("=======================")
			r.infoLog.Println("Command Log:")
			fmt.Println("\n" + string(b))
			r.infoLog.Println("=======================")
			color.Unset()
		}
	}

	return nil
}

func (r Runner) internalExec(command string) error {
	args := strings.Split(command, " ")

	exec, err := r.cli.ContainerExecCreate(r.ctx, r.container.ID, types.ExecConfig{
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

	_, err = r.cli.ContainerExecInspect(r.ctx, exec.ID)
	if err != nil {
		return err
	}

	return nil
}
