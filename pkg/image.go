package pin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/fatih/color"
)

type ImagePullingResult struct {
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

func (r runner) checkTheImageAvailable() (bool, error) {
	images, err := r.cli.ImageList(r.ctx, types.ImageListOptions{})

	if err != nil {
		return false, err
	}

	for _, v := range images {
		if r.currentJob.Image == v.RepoTags[0] {
			color.Set(color.FgGreen)
			r.infoLog.Println("Image is available")
			color.Unset()
			return true, nil
		}
	}

	return false, nil
}

func (r runner) pullImage() error {
	color.Set(color.FgBlue)
	r.infoLog.Printf("Image pulling: %s", r.currentJob.Image)
	color.Unset()

	r.infoLog.Println("Waiting for docker response...")

	reader, err := r.cli.ImagePull(r.ctx, r.currentJob.Image, types.ImagePullOptions{})

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

		imagePullingResult := ImagePullingResult{}

		err = json.Unmarshal([]byte(sline), &imagePullingResult)

		if err != nil {
			return err
		}

		fmt.Printf("\033[A\033[K%s %s\n", imagePullingResult.Status, imagePullingResult.Progress)
	}

	return nil
}
