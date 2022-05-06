package pin

import (
	"archive/tar"
	"bytes"
	"os"

	"github.com/docker/docker/api/types"
)

func (r *runner) prepareAndRunShellCommandScript() error {
	if r.currentJob.SoloExecution {
		for _, cmd := range r.currentJob.Script {
			err := r.commandScriptExecutor(cmd)

			if err != nil {
				return err
			}
		}
	} else {
		userCommandLines := ""

		for _, cmd := range r.currentJob.Script {
			userCommandLines += cmd + "\n"
		}

		err := r.commandScriptExecutor(userCommandLines)

		if err != nil {
			return err
		}
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

	err = r.cli.CopyToContainer(r.ctx, r.containerResponse.ID, "/home/", &buf, types.CopyToContainerOptions{})

	if err != nil {
		return err
	}

	return nil
}
