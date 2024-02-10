package runner

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/internal/container_manager"
	"github.com/muhammedikinci/pin/internal/image_manager"
	"github.com/muhammedikinci/pin/internal/interfaces"
	"github.com/muhammedikinci/pin/internal/shell_commander"
)

type Runner struct {
	ctx context.Context
	cli interfaces.Client
}

func (r *Runner) run(pipeline Pipeline) error {
	r.createGlobalContext(pipeline.Workflow)

	cli, err := client.NewClientWithOpts()

	if err != nil {
		return err
	}

	r.cli = cli

	for _, job := range pipeline.Workflow {
		go func(job *Job) {
			r.jobRunner(job, pipeline.LogsWithTime)
		}(job)
	}

	err = <-pipeline.Workflow[len(pipeline.Workflow)-1].ErrorChannel

	return err
}

func (r *Runner) jobRunner(currentJob *Job, logsWithTime bool) {
	if logsWithTime {
		currentJob.InfoLog = log.New(os.Stdout, fmt.Sprintf("⚉ %s ", currentJob.Name), log.Ldate|log.Ltime)
	} else {
		currentJob.InfoLog = log.New(os.Stdout, fmt.Sprintf("⚉ %s ", currentJob.Name), 0)
	}

	currentJob.ImageManager = image_manager.NewImageManager(r.cli, currentJob.InfoLog)
	currentJob.ContainerManager = container_manager.NewContainerManager(r.cli, currentJob.InfoLog)
	currentJob.ShellCommander = shell_commander.NewShellCommander()

	if currentJob.Previous != nil && !currentJob.IsParallel {
		previousJobError := <-currentJob.Previous.ErrorChannel

		if previousJobError != nil {
			currentJob.ErrorChannel <- nil
			return
		}
	}

	isImageAvailable, err := currentJob.ImageManager.CheckTheImageAvailable(r.ctx, currentJob.Image)

	if err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if !isImageAvailable {
		if err := currentJob.ImageManager.PullImage(r.ctx, currentJob.Image); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
	}

	ports := map[string]string{}

	for _, port := range currentJob.Port {
		ports[port.Out] = port.In
	}

	resp, err := currentJob.ContainerManager.StartContainer(r.ctx, currentJob.Name, currentJob.Image, ports)

	if err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	currentJob.Container = resp

	if currentJob.CopyFiles {
		if err := currentJob.ContainerManager.CopyToContainer(r.ctx, resp.ID, currentJob.WorkDir, currentJob.CopyIgnore); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
	}

	color.Set(color.FgGreen)
	currentJob.InfoLog.Println("Starting the container")
	color.Unset()

	if err := r.cli.ContainerStart(r.ctx, currentJob.Container.ID, types.ContainerStartOptions{}); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := r.commandScriptExecutor((*currentJob)); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := currentJob.ContainerManager.StopContainer(r.ctx, currentJob.Container.ID); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := currentJob.ContainerManager.RemoveContainer(r.ctx, currentJob.Container.ID, false); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	color.Set(color.FgGreen)
	currentJob.InfoLog.Println("Job ended")
	color.Unset()

	currentJob.ErrorChannel <- nil
}

func (r Runner) commandScriptExecutor(currentJob Job) error {
	cmds := currentJob.ShellCommander.PrepareShellCommands(currentJob.SoloExecution, currentJob.Script)

	for _, cmd := range cmds {
		buf, err := currentJob.ShellCommander.ShellToTar(cmd)

		if err != nil {
			return err
		}

		err = r.cli.CopyToContainer(r.ctx, currentJob.Container.ID, "/home/", buf, types.CopyToContainerOptions{})

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
		currentJob.InfoLog.Printf("Execute command: %s", name)
	} else if !currentJob.SoloExecution {
		currentJob.InfoLog.Println("soloExecution disabled, shell command started!")
	}

	exec, err := r.cli.ContainerExecCreate(r.ctx, currentJob.Container.ID, types.ExecConfig{
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
		currentJob.InfoLog.Printf("Command execution failed")

		currentJob.InfoLog.Println("Command Log:")

		if reader, _, err := r.cli.CopyFromContainer(r.ctx, currentJob.Container.ID, "/shell_command_output.log"); err == nil {
			tr := tar.NewReader(reader)
			tr.Next()
			b, _ := io.ReadAll(tr)
			fmt.Println("\n" + string(b))
		}
		color.Unset()

		r.cli.ContainerKill(r.ctx, currentJob.Container.ID, "KILL")

		if err := currentJob.ContainerManager.StopContainer(r.ctx, currentJob.Container.ID); err != nil {
			return err
		}

		if err := currentJob.ContainerManager.RemoveContainer(r.ctx, currentJob.Container.ID, false); err != nil {
			return err
		}

		return errors.New("command execution failed")
	}

	currentJob.InfoLog.Println("Command execution successful")

	if reader, _, err := r.cli.CopyFromContainer(r.ctx, currentJob.Container.ID, "/shell_command_output.log"); err == nil {
		tr := tar.NewReader(reader)
		tr.Next()
		b, _ := io.ReadAll(tr)

		if len(b) != 0 {
			color.Set(color.FgGreen)
			currentJob.InfoLog.Println("Command Log:")
			fmt.Println("\n" + string(b))
			color.Unset()
		}
	}

	return nil
}

func (r Runner) internalExec(command string, currentJob Job) error {
	args := strings.Split(command, " ")

	exec, err := r.cli.ContainerExecCreate(r.ctx, currentJob.Container.ID, types.ExecConfig{
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

func (r *Runner) createGlobalContext(jobs []*Job) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go func() {
		<-ctx.Done()
		color.Set(color.FgHiRed)
		fmt.Println("System call detected!")
		color.Unset()
		cancel()

		for _, job := range jobs {
			if job.Container.ID == "" {
				continue
			}

			timedContext, timedCancel := context.WithTimeout(context.Background(), time.Second*3)
			defer timedCancel()
			job.ContainerManager.RemoveContainer(timedContext, job.Container.ID, true)
		}
	}()

	r.ctx = ctx
}
