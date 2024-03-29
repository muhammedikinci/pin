package image_manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/muhammedikinci/pin/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWhenImageListReturnAnyErrorCheckTheImageAvailableMustReturnFalseAndCliError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	merr := errors.New("test")
	mimages := []types.ImageSummary{}

	mockCli.
		EXPECT().
		ImageList(gomock.Any(), gomock.Any()).
		Return(mimages, merr)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	check, err := im.CheckTheImageAvailable(context.Background(), "test")

	assert.Equal(t, err, merr)
	assert.Equal(t, check, false)
}

func TestWhenCheckTheImageAvailableCallsWithDoesntExistImageMustReturnFalseAndNilError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mimages := []types.ImageSummary{
		{
			RepoTags: []string{"asd"},
		},
	}

	mockCli.
		EXPECT().
		ImageList(gomock.Any(), gomock.Any()).
		Return(mimages, nil)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	check, err := im.CheckTheImageAvailable(context.Background(), "test")

	assert.Equal(t, err, nil)
	assert.Equal(t, check, false)
}

func TestWhenCheckTheImageAvailableCallsWithExistImageMustReturnTrueAndNilError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mimages := []types.ImageSummary{
		{
			RepoTags: []string{"image1"},
		},
	}

	mockCli.
		EXPECT().
		ImageList(gomock.Any(), gomock.Any()).
		Return(mimages, nil)

	mockLog.EXPECT().
		Println("Image is available").
		Times(1)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	check, err := im.CheckTheImageAvailable(context.Background(), "image1")

	assert.Equal(t, err, nil)
	assert.Equal(t, check, true)
}

func TestWhenClientImagePullFunctionReturnAnErrorPullImageMustReturnTheSameError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mimage := "test"
	merr := errors.New("test")
	stringReadCloser := io.NopCloser(nil)

	mockCli.
		EXPECT().
		ImagePull(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(stringReadCloser, merr)

	mockLog.EXPECT().
		Println("Waiting for docker response...").
		Times(1)
	mockLog.EXPECT().
		Printf("Image pulling: %s", mimage).
		Times(1)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	err := im.PullImage(context.Background(), mimage)

	assert.Equal(t, err, merr)
}

func TestWhenClientImagePullFunctionReturnUnexpectedStreamPullImageMustReturnTheUnmarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mimage := "test"
	var buf bytes.Buffer

	fmt.Fprintln(&buf, `{"status": "test1", prog`)

	bufreader := bytes.NewReader(buf.Bytes())
	reader := io.NopCloser(bufreader)

	mockCli.
		EXPECT().
		ImagePull(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(reader, nil)

	mockLog.EXPECT().
		Println("Waiting for docker response...").
		Times(1)
	mockLog.EXPECT().
		Printf("Image pulling: %s", mimage).
		Times(1)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	err := im.PullImage(context.Background(), mimage)

	var want *json.SyntaxError

	assert.Equal(t, errors.As(err, &want), true)
}

func TestWhenClientImagePullFunctionReturnSuccessfulStreamPullImageMustReturnNilAndPrintLogs(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	mockCli := mocks.NewMockClient(ctrl)
	mockLog := mocks.NewMockLog(ctrl)

	mimage := "test"
	var buf bytes.Buffer

	fmt.Fprintln(&buf, `{"status": "test1"}`)

	reader := io.NopCloser(strings.NewReader(buf.String()))

	mockCli.
		EXPECT().
		ImagePull(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(reader, nil)

	mockLog.EXPECT().
		Println("Waiting for docker response...").
		Times(1)
	mockLog.EXPECT().
		Printf("Image pulling: %s", mimage).
		Times(1)

	im := imageManager{
		cli: mockCli,
		log: mockLog,
	}

	err := im.PullImage(context.Background(), mimage)

	assert.Equal(t, err, nil)
}
