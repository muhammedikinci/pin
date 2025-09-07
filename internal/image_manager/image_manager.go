package image_manager

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/fatih/color"
	"github.com/muhammedikinci/pin/internal/interfaces"
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
	images, err := im.cli.ImageList(ctx, imagetypes.ListOptions{})

	if err != nil {
		return false, err
	}

	for _, v := range images {
		for _, tag := range v.RepoTags {
			if image == tag {
				color.Set(color.FgGreen)
				im.log.Println("Image is available")
				color.Unset()
				return true, nil
			}
		}
	}

	return false, nil
}

func (im imageManager) PullImage(ctx context.Context, image string) error {
	color.Set(color.FgBlue)
	im.log.Printf("Image pulling: %s", image)
	color.Unset()

	im.log.Println("Waiting for docker response...")

	reader, err := im.cli.ImagePull(ctx, image, imagetypes.PullOptions{})

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

func (im imageManager) BuildImageFromDockerfile(ctx context.Context, dockerfilePath string, imageName string) error {
	color.Set(color.FgBlue)
	im.log.Printf("Building image from Dockerfile: %s", dockerfilePath)
	color.Unset()

	// Create a tar archive containing the Dockerfile and build context
	buf, err := im.createDockerfileTar(dockerfilePath)
	if err != nil {
		return err
	}

	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{imageName},
		Remove:     true,
		Context:    buf,
	}

	buildResponse, err := im.cli.ImageBuild(ctx, buf, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	// Read and display build output
	scanner := bufio.NewScanner(buildResponse.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var buildResult map[string]interface{}
		if err := json.Unmarshal([]byte(line), &buildResult); err == nil {
			if stream, ok := buildResult["stream"].(string); ok {
				fmt.Print(strings.TrimSuffix(stream, "\n"))
			}
			if errorDetail, ok := buildResult["errorDetail"].(map[string]interface{}); ok {
				if message, ok := errorDetail["message"].(string); ok {
					color.Set(color.FgRed)
					im.log.Printf("Build error: %s", message)
					color.Unset()
					return fmt.Errorf("docker build failed: %s", message)
				}
			}
		}
	}

	color.Set(color.FgGreen)
	im.log.Printf("Image built successfully: %s", imageName)
	color.Unset()

	return nil
}

func (im imageManager) createDockerfileTar(dockerfilePath string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Get the directory containing the Dockerfile for build context
	dockerfileDir := filepath.Dir(dockerfilePath)
	
	// Walk through the build context directory
	err := filepath.Walk(dockerfileDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// Update the name to be relative to the build context
		relPath, err := filepath.Rel(dockerfileDir, file)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// Special handling for the Dockerfile
		if filepath.Base(file) == filepath.Base(dockerfilePath) {
			header.Name = "Dockerfile"
		}

		// Write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, write the content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			defer data.Close()

			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return buf, nil
}
