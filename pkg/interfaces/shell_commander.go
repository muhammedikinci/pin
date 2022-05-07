package interfaces

import "bytes"

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type ShellCommander interface {
	PrepareShellCommands(soloExecution bool, scripts []string) []string
	ShellToTar(cmd string) (*bytes.Buffer, error)
}
