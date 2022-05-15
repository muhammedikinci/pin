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
	"os/signal"
	"strings"
	"syscall"

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
	infoLog          *log.Logger
	imageManager     interfaces.ImageManager
	containerManager interfaces.ContainerManager
	shellCommander   interfaces.ShellCommander
	container        container.ContainerCreateCreatedBody
}

func (r *Runner) run(pipeline Pipeline) error {
	if pipeline.LogsWithTime {
		r.infoLog = log.New(os.Stdout, "⚉ ", log.Ldate|log.Ltime)
	} else {
		r.infoLog = log.New(os.Stdout, "⚉ ", 0)
	}

	r.createGlobalContext()

	cli, err := client.NewClientWithOpts()

	if err != nil {
		return err
	}

	r.cli = cli
	r.imageManager = image_manager.NewImageManager(r.cli, r.infoLog)
	r.containerManager = container_manager.NewContainerManager(r.cli, r.infoLog)
	r.shellCommander = shell_commander.NewShellCommander()

	for _, job := range pipeline.Workflow {
		go func(job Job) {
			r.jobRunner(job)
		}(job)
	}

	err = <-pipeline.Workflow[len(pipeline.Workflow)-1].ErrorChannel

	return err
}

func (r *Runner) jobRunner(currentJob Job) {
	if currentJob.Previous != nil {
		previousJobError := <-currentJob.Previous.ErrorChannel

		if previousJobError != nil {
			currentJob.ErrorChannel <- nil
			return
		}
	}

	isImageAvailable, err := r.imageManager.CheckTheImageAvailable(r.ctx, currentJob.Image)

	if err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if !isImageAvailable {
		if err := r.imageManager.PullImage(r.ctx, currentJob.Image); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
	}

	ports := map[string]string{}

	for _, port := range currentJob.Port {
		ports[port.Out] = port.In
	}

	resp, err := r.containerManager.StartContainer(r.ctx, currentJob.Name, currentJob.Image, ports)

	if err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	r.container = resp

	if currentJob.CopyFiles {
		if err := r.containerManager.CopyToContainer(r.ctx, resp.ID, currentJob.WorkDir, currentJob.CopyIgnore); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
	}

	color.Set(color.FgGreen)
	r.infoLog.Println("Starting the container")
	color.Unset()

	if err := r.cli.ContainerStart(r.ctx, r.container.ID, types.ContainerStartOptions{}); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := r.commandScriptExecutor(currentJob); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := r.containerManager.StopContainer(r.ctx, r.container.ID); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := r.containerManager.RemoveContainer(r.ctx, r.container.ID, false); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	color.Set(color.FgGreen)
	r.infoLog.Println("Job ended")
	color.Unset()

	currentJob.ErrorChannel <- nil
}

func (r Runner) commandScriptExecutor(currentJob Job) error {
	cmds := r.shellCommander.PrepareShellCommands(currentJob.SoloExecution, currentJob.Script)

	for _, cmd := range cmds {
		buf, err := r.shellCommander.ShellToTar(cmd)

		if err != nil {
			return err
		}

		err = r.cli.CopyToContainer(r.ctx, r.container.ID, "/home/", buf, types.CopyToContainerOptions{})

		if err != nil {
			return err
		}

		if err := r.internalExec("chmod +x /home/shell_command.sh", currentJob); err != nil {
			return err
		}

		if err := r.commandRunner("sh /home/shell_command.sh", cmd, currentJob); err != nil {
			return err
		}

		if err := r.internalExec("rm /home/shell_command.sh", currentJob); err != nil {
			return err
		}
	}

	return nil
}

func (r Runner) commandRunner(command string, name string, currentJob Job) error {
	args := strings.Split(command, " ")

	if name != "" && currentJob.SoloExecution {
		lines := strings.Split(name, "\n")
		name = strings.Join(lines[2:], "\n")
		r.infoLog.Printf("Execute command: %s", name)
	} else if !currentJob.SoloExecution {
		r.infoLog.Println("soloExecution disabled, shell command started!")
	}

	exec, err := r.cli.ContainerExecCreate(r.ctx, r.container.ID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          args,
		WorkingDir:   currentJob.WorkDir,
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

		r.infoLog.Println("Command Log:")

		if reader, _, err := r.cli.CopyFromContainer(r.ctx, r.container.ID, "/shell_command_output.log"); err == nil {
			tr := tar.NewReader(reader)
			tr.Next()
			b, _ := ioutil.ReadAll(tr)
			fmt.Println("\n" + string(b))
		}
		color.Unset()

		r.cli.ContainerKill(r.ctx, r.container.ID, "KILL")

		if err := r.containerManager.StopContainer(r.ctx, r.container.ID); err != nil {
			return err
		}

		if err := r.containerManager.RemoveContainer(r.ctx, r.container.ID, false); err != nil {
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
			r.infoLog.Println("Command Log:")
			fmt.Println("\n" + string(b))
			color.Unset()
		}
	}

	return nil
}

func (r Runner) internalExec(command string, currentJob Job) error {
	args := strings.Split(command, " ")

	exec, err := r.cli.ContainerExecCreate(r.ctx, r.container.ID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          args,
		WorkingDir:   currentJob.WorkDir,
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

func (r *Runner) createGlobalContext() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		color.Set(color.FgHiRed)
		r.infoLog.Println("System call detected!")
		color.Unset()
		cancel()

		if r.container.ID == "" {
			return
		}

		r.containerManager.RemoveContainer(context.Background(), r.container.ID, true)
	}()

	r.ctx = ctx
}
