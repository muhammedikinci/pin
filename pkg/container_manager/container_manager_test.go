package container_manager

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		cli: mockCli,
		log: mockLog,
	}

	resp, err := cm.StartContainer(context.Background(), "", "", map[string]string{})

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
		cli: mockCli,
		log: mockLog,
	}

	resp, err := cm.StartContainer(context.Background(), "", "", map[string]string{})

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
		cli: mockCli,
		log: mockLog,
	}

	err := cm.StopContainer(context.Background(), "")

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
		cli: mockCli,
		log: mockLog,
	}

	err := cm.StopContainer(context.Background(), "")

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
		cli: mockCli,
		log: mockLog,
	}

	err := cm.RemoveContainer(context.Background(), "", false)

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
		cli: mockCli,
		log: mockLog,
	}

	err := cm.RemoveContainer(context.Background(), "", false)

	assert.Equal(t, err, nil)
}

func TestAppender(t *testing.T) {
	var buf bytes.Buffer

	tw := tar.NewWriter(&buf)
	basepath, _ := os.Getwd()
	dir := filepath.Dir(filepath.Dir(basepath))
	currentPath := filepath.FromSlash(path.Join(dir, "testdata"))
	fmt.Println(dir)
	cm := containerManager{}

	err := filepath.Walk(currentPath, func(path string, info os.FileInfo, err error) error {
		return cm.appender(path, info, err, currentPath, tw, []string{"node_modules", "ignore_test1.txt", ".test_point_folder"})
	})

	assert.Equal(t, err, nil)

	tw.Close()

	tr := tar.NewReader(&buf)

	headerNames := []string{}

	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		if header == nil {
			break
		}

		assert.Equal(t, false, strings.Contains(header.Name, ".test_point_folder"))
		assert.Equal(t, false, strings.Contains(header.Name, "node_modules"))
		assert.Equal(t, false, strings.Contains(header.Name, "ignore_test1.txt"))

		headerNames = append(headerNames, header.Name)
	}

	assert.Contains(t, headerNames, "ignore_test/ignore_test2.py")
}
