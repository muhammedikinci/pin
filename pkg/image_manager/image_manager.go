package image_manager

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/pkg/interfaces"
)

func NewImageManager(cli interfaces.Client, log interfaces.Log) imageManager {
	return imageManager{
		cli: cli,
		log: log,
	}
}

type imageManager struct {
	cli interfaces.Client
	log interfaces.Log
}

type imagePullingResult struct {
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

func (im imageManager) CheckTheImageAvailable(ctx context.Context, image string) (bool, error) {
	images, err := im.cli.ImageList(ctx, types.ImageListOptions{})

	if err != nil {
		return false, err
	}

	for _, v := range images {
		if image == v.RepoTags[0] {
			color.Set(color.FgGreen)
			im.log.Println("Image is available")
			color.Unset()
			return true, nil
		}
	}

	return false, nil
}

func (im imageManager) PullImage(ctx context.Context, image string) error {
	color.Set(color.FgBlue)
	im.log.Printf("Image pulling: %s", image)
	color.Unset()

	im.log.Println("Waiting for docker response...")

	reader, err := im.cli.ImagePull(ctx, image, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	defer reader.Close()

	bio := bufio.NewReader(reader)

	for {
		line, err := bio.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		sline := strings.TrimRight(string(line), "\n")

		res := imagePullingResult{}

		err = json.Unmarshal([]byte(sline), &res)

		if err != nil {
			return err
		}

		fmt.Printf("\033[A\033[K%s %s\n", res.Status, res.Progress)
	}

	return nil
}
