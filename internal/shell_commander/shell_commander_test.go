package shell_commander

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type prepareShellCommandsTestCase struct {
	soloExecution bool
	scripts       []string
	result        []string
}

func TestPrepareShellCommands(t *testing.T) {
	shellCommander := NewShellCommander()

	testCases := []prepareShellCommandsTestCase{
		{
			soloExecution: true,
			scripts: []string{
				"go test ./...",
			},
			result: []string{
				"#!/bin/sh\nexec > /shell_command_output.log 2>&1\ngo test ./...",
			},
		},
		{
			soloExecution: true,
			scripts: []string{
				"go test ./...",
				"npm install",
			},
			result: []string{
				"#!/bin/sh\nexec > /shell_command_output.log 2>&1\ngo test ./...",
				"#!/bin/sh\nexec > /shell_command_output.log 2>&1\nnpm install",
			},
		},
		{
			soloExecution: false,
			scripts: []string{
				"go test ./...",
				"npm install",
			},
			result: []string{
				"#!/bin/sh\nexec > /shell_command_output.log 2>&1\ngo test ./...\nnpm install\n",
			},
		},
		{
			soloExecution: true,
			scripts:       []string{},
			result:        []string{},
		},
		{
			soloExecution: false,
			scripts:       []string{},
			result:        []string{},
		},
	}

	for _, testCase := range testCases {
		res := shellCommander.PrepareShellCommands(testCase.soloExecution, testCase.scripts)

		assert.Equal(t, len(res), len(testCase.result))
		assert.Equal(t, res, testCase.result)
	}
}
