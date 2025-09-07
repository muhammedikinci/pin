package runner

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
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

	// Create Docker client with custom host if specified
	var cli interfaces.Client
	var err error
	
	if pipeline.DockerHost != "" {
		cli, err = client.NewClientWithOpts(
			client.WithHost(pipeline.DockerHost),
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
	} else {
		cli, err = client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
	}
	
	if err != nil {
		return err
	}

	r.cli = cli

	for _, job := range pipeline.Workflow {
		go func(job *Job) {
			r.jobRunnerWithRetry(job, pipeline.LogsWithTime)
		}(job)
	}

	err = <-pipeline.Workflow[len(pipeline.Workflow)-1].ErrorChannel

	return err
}

// jobRunnerWithRetry handles job execution with retry logic
func (r *Runner) jobRunnerWithRetry(currentJob *Job, logsWithTime bool) {
	// Set up logging first
	if logsWithTime {
		currentJob.InfoLog = log.New(
			os.Stdout,
			fmt.Sprintf("⚉ %s ", currentJob.Name),
			log.Ldate|log.Ltime,
		)
	} else {
		currentJob.InfoLog = log.New(os.Stdout, fmt.Sprintf("⚉ %s ", currentJob.Name), 0)
	}
	
	var lastError error
	
	for attempt := 1; attempt <= currentJob.RetryConfig.MaxAttempts; attempt++ {
		// Create a copy of the job for this attempt to reset state
		attemptJob := *currentJob
		attemptJob.ErrorChannel = make(chan error, 1)
		
		// Run the job attempt
		go r.jobRunner(&attemptJob, logsWithTime)
		lastError = <-attemptJob.ErrorChannel
		
		if lastError == nil {
			// Success - send success to original error channel
			currentJob.ErrorChannel <- nil
			return
		}
		
		// If this was the last attempt, send the error
		if attempt == currentJob.RetryConfig.MaxAttempts {
			color.Set(color.FgRed)
			currentJob.InfoLog.Printf("Job failed after %d attempts", currentJob.RetryConfig.MaxAttempts)
			color.Unset()
			currentJob.ErrorChannel <- lastError
			return
		}
		
		// Calculate delay with exponential backoff
		delay := time.Duration(float64(currentJob.RetryConfig.DelaySeconds) * math.Pow(currentJob.RetryConfig.BackoffMultiplier, float64(attempt-1))) * time.Second
		
		color.Set(color.FgYellow)
		currentJob.InfoLog.Printf("Job failed (attempt %d/%d), retrying in %v: %s", 
			attempt, currentJob.RetryConfig.MaxAttempts, delay, lastError.Error())
		color.Unset()
		
		// Wait before retrying
		time.Sleep(delay)
	}
}

func (r *Runner) jobRunner(currentJob *Job, logsWithTime bool) {
	if logsWithTime {
		currentJob.InfoLog = log.New(
			os.Stdout,
			fmt.Sprintf("⚉ %s ", currentJob.Name),
			log.Ldate|log.Ltime,
		)
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

	conditionEvaluator := NewConditionEvaluator()
	if currentJob.Condition != "" && !conditionEvaluator.EvaluateCondition(currentJob.Condition) {
		color.Set(color.FgYellow)
		currentJob.InfoLog.Printf("Job skipped due to condition: %s", currentJob.Condition)
		color.Unset()
		currentJob.ErrorChannel <- nil
		return
	}

	// Handle Dockerfile build or regular image pull/check
	if currentJob.Dockerfile != "" {
		// Build image from Dockerfile
		imageName := fmt.Sprintf("%s-custom:%s", currentJob.Name, "latest")
		if err := currentJob.ImageManager.BuildImageFromDockerfile(r.ctx, currentJob.Dockerfile, imageName); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
		// Use the built image name
		currentJob.Image = imageName
	} else if currentJob.Image != "" {
		// Handle regular image pull/check
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
	} else {
		currentJob.ErrorChannel <- errors.New("either 'image' or 'dockerfile' must be specified")
		return
	}

	ports := map[string]string{}

	for _, port := range currentJob.Port {
		// Create host info with IP if specified
		var hostInfo string
		if port.HostIP != "" && port.HostIP != "0.0.0.0" {
			hostInfo = port.HostIP + ":" + port.Out
		} else {
			hostInfo = port.Out
		}
		ports[hostInfo] = port.In
	}

	resp, err := currentJob.ContainerManager.StartContainer(
		r.ctx,
		currentJob.Name,
		currentJob.Image,
		ports,
		currentJob.Env,
	)
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

	if err := r.cli.ContainerStart(r.ctx, currentJob.Container.ID, container.StartOptions{}); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if err := r.commandScriptExecutor((*currentJob)); err != nil {
		currentJob.ErrorChannel <- err
		return
	}

	if currentJob.ArtifactPath != "" {
		if err := currentJob.ContainerManager.CopyFromContainer(r.ctx, currentJob.Container.ID, currentJob.ArtifactPath, "./*"); err != nil {
			currentJob.ErrorChannel <- err
			return
		}
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
	cmds := currentJob.ShellCommander.PrepareShellCommands(
		currentJob.SoloExecution,
		currentJob.Script,
	)

	for _, cmd := range cmds {
		buf, err := currentJob.ShellCommander.ShellToTar(cmd)
		if err != nil {
			return err
		}

		err = r.cli.CopyToContainer(
			r.ctx,
			currentJob.Container.ID,
			"/home/",
			buf,
			container.CopyToContainerOptions{},
		)
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

	exec, err := r.cli.ContainerExecCreate(r.ctx, currentJob.Container.ID, container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          args,
		WorkingDir:   currentJob.WorkDir,
	})
	if err != nil {
		return err
	}

	res, err := r.cli.ContainerExecAttach(r.ctx, exec.ID, container.ExecAttachOptions{Tty: true})
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

	exec, err := r.cli.ContainerExecCreate(r.ctx, currentJob.Container.ID, container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          args,
		WorkingDir:   currentJob.WorkDir,
	})
	if err != nil {
		return err
	}

	res, err := r.cli.ContainerExecAttach(r.ctx, exec.ID, container.ExecAttachOptions{Tty: true})
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
