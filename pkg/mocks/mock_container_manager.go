// Code generated by MockGen. DO NOT EDIT.
// Source: container_manager.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	container "github.com/docker/docker/api/types/container"
	gomock "github.com/golang/mock/gomock"
)

// MockContainerManager is a mock of ContainerManager interface.
type MockContainerManager struct {
	ctrl     *gomock.Controller
	recorder *MockContainerManagerMockRecorder
}

// MockContainerManagerMockRecorder is the mock recorder for MockContainerManager.
type MockContainerManagerMockRecorder struct {
	mock *MockContainerManager
}

// NewMockContainerManager creates a new mock instance.
func NewMockContainerManager(ctrl *gomock.Controller) *MockContainerManager {
	mock := &MockContainerManager{ctrl: ctrl}
	mock.recorder = &MockContainerManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContainerManager) EXPECT() *MockContainerManagerMockRecorder {
	return m.recorder
}

// CopyToContainer mocks base method.
func (m *MockContainerManager) CopyToContainer(ctx context.Context, containerID, workDir string, copyIgnore []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CopyToContainer", ctx, containerID, workDir, copyIgnore)
	ret0, _ := ret[0].(error)
	return ret0
}

// CopyToContainer indicates an expected call of CopyToContainer.
func (mr *MockContainerManagerMockRecorder) CopyToContainer(ctx, containerID, workDir, copyIgnore interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CopyToContainer", reflect.TypeOf((*MockContainerManager)(nil).CopyToContainer), ctx, containerID, workDir, copyIgnore)
}

// RemoveContainer mocks base method.
func (m *MockContainerManager) RemoveContainer(ctx context.Context, containerID string, forceRemove bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveContainer", ctx, containerID, forceRemove)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveContainer indicates an expected call of RemoveContainer.
func (mr *MockContainerManagerMockRecorder) RemoveContainer(ctx, containerID, forceRemove interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveContainer", reflect.TypeOf((*MockContainerManager)(nil).RemoveContainer), ctx, containerID, forceRemove)
}

// StartContainer mocks base method.
func (m *MockContainerManager) StartContainer(ctx context.Context, jobName, image string, ports map[string]string) (container.ContainerCreateCreatedBody, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartContainer", ctx, jobName, image, ports)
	ret0, _ := ret[0].(container.ContainerCreateCreatedBody)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StartContainer indicates an expected call of StartContainer.
func (mr *MockContainerManagerMockRecorder) StartContainer(ctx, jobName, image, ports interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartContainer", reflect.TypeOf((*MockContainerManager)(nil).StartContainer), ctx, jobName, image, ports)
}

// StopContainer mocks base method.
func (m *MockContainerManager) StopContainer(ctx context.Context, containerID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StopContainer", ctx, containerID)
	ret0, _ := ret[0].(error)
	return ret0
}

// StopContainer indicates an expected call of StopContainer.
func (mr *MockContainerManagerMockRecorder) StopContainer(ctx, containerID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopContainer", reflect.TypeOf((*MockContainerManager)(nil).StopContainer), ctx, containerID)
}
