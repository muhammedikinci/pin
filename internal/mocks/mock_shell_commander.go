// Code generated by MockGen. DO NOT EDIT.
// Source: shell_commander.go

// Package mocks is a generated GoMock package.
package mocks

import (
	bytes "bytes"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockShellCommander is a mock of ShellCommander interface.
type MockShellCommander struct {
	ctrl     *gomock.Controller
	recorder *MockShellCommanderMockRecorder
}

// MockShellCommanderMockRecorder is the mock recorder for MockShellCommander.
type MockShellCommanderMockRecorder struct {
	mock *MockShellCommander
}

// NewMockShellCommander creates a new mock instance.
func NewMockShellCommander(ctrl *gomock.Controller) *MockShellCommander {
	mock := &MockShellCommander{ctrl: ctrl}
	mock.recorder = &MockShellCommanderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockShellCommander) EXPECT() *MockShellCommanderMockRecorder {
	return m.recorder
}

// PrepareShellCommands mocks base method.
func (m *MockShellCommander) PrepareShellCommands(soloExecution bool, scripts []string) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrepareShellCommands", soloExecution, scripts)
	ret0, _ := ret[0].([]string)
	return ret0
}

// PrepareShellCommands indicates an expected call of PrepareShellCommands.
func (mr *MockShellCommanderMockRecorder) PrepareShellCommands(soloExecution, scripts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrepareShellCommands", reflect.TypeOf((*MockShellCommander)(nil).PrepareShellCommands), soloExecution, scripts)
}

// ShellToTar mocks base method.
func (m *MockShellCommander) ShellToTar(cmd string) (*bytes.Buffer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShellToTar", cmd)
	ret0, _ := ret[0].(*bytes.Buffer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShellToTar indicates an expected call of ShellToTar.
func (mr *MockShellCommanderMockRecorder) ShellToTar(cmd interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShellToTar", reflect.TypeOf((*MockShellCommander)(nil).ShellToTar), cmd)
}