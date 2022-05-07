package shell_commander

import (
	"archive/tar"
	"bytes"
)

type ShellCommander struct {
}

func NewShellCommander() ShellCommander {
	return ShellCommander{}
}

func (sc ShellCommander) PrepareShellCommands(soloExecution bool, scripts []string) []string {
	cmds := []string{}

	if soloExecution {
		for _, cmd := range scripts {
			cmds = append(cmds, sc.wrapCommand(cmd))
		}
	} else {
		userCommandLines := ""

		for _, cmd := range scripts {
			userCommandLines += cmd + "\n"
		}

		cmds = append(cmds, sc.wrapCommand(userCommandLines))
	}

	return cmds
}

func (sc ShellCommander) wrapCommand(cmd string) string {
	return "#!/bin/sh\nexec > /shell_command_output.log 2>&1\n" + cmd
}

func (sc ShellCommander) ShellToTar(cmd string) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	defer tw.Close()

	err := tw.WriteHeader(&tar.Header{
		Name: "shell_command.sh",
		Mode: 0777,
		Size: int64(len(cmd)),
	})

	if err != nil {
		return &buf, err
	}

	_, err = tw.Write([]byte(cmd))

	if err != nil {
		return &buf, err
	}

	return &buf, nil
}
