package container_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
	"github.com/muhammedikinci/pin/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWhenContainerCreateReturnErrorStartContainerMustReturnSameError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	merror := errors.New("test")

	mockLog.
		EXPECT().
		Println("Start creating container")

	mockCli.
		EXPECT().
		ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(container.ContainerCreateCreatedBody{}, merror)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	resp, err := cm.StartContainer("", "")

	assert.Equal(t, resp, container.ContainerCreateCreatedBody{})
	assert.Equal(t, err, merror)
}

func TestWhenContainerCreateReturnResponseStartContainerMustSameResponseWithNilError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mres := container.ContainerCreateCreatedBody{
		ID: "test",
	}

	mockLog.
		EXPECT().
		Println("Start creating container")

	mockCli.
		EXPECT().
		ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mres, nil)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	resp, err := cm.StartContainer("", "")

	assert.Equal(t, resp.ID, mres.ID)
	assert.Equal(t, err, nil)
}

func TestWhenContainerStopReturnErrorStopContainerMustReturnSameError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	merror := errors.New("test")

	mockLog.
		EXPECT().
		Println("Container stopping")

	mockCli.
		EXPECT().
		ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(merror)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	err := cm.StopContainer("")

	assert.Equal(t, err, merror)
}

func TestWhenContainerStopReturnNilStopContainerMustReturnNil(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mockLog.
		EXPECT().
		Println("Container stopping")

	mockLog.
		EXPECT().
		Println("Container stopped")

	mockCli.
		EXPECT().
		ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	err := cm.StopContainer("")

	assert.Equal(t, err, nil)
}

func TestWhenRemoveContainerReturnErrorStopContainerMustReturnSameError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	merror := errors.New("test")

	mockLog.
		EXPECT().
		Println("Container removing")

	mockCli.
		EXPECT().
		ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(merror)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	err := cm.RemoveContainer("")

	assert.Equal(t, err, merror)
}

func TestWhenContainerRemoveReturnNilRemoveContainerMustReturnNil(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mockLog.
		EXPECT().
		Println("Container removing")

	mockLog.
		EXPECT().
		Println("Container removed")

	mockCli.
		EXPECT().
		ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	cm := containerManager{
		ctx: context.Background(),
		cli: mockCli,
		log: mockLog,
	}

	err := cm.RemoveContainer("")

	assert.Equal(t, err, nil)
}
