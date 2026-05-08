package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const defaultPipelineTemplate = `workflow:
  - build
  - test
  - deploy

logsWithTime: true

build:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go mod download
    - go build ./...

test:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test ./...

deploy:
  host: true
  condition: $BRANCH == "main"
  script:
    - echo "Deploy step goes here"
`

var initFilePath string
var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a starter pin.yaml",
	Long:  "Create a starter pipeline file for build, test, and deploy jobs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if initFilePath == "" {
			return fmt.Errorf("required flag \"filepath\" not set")
		}

		if !initForce {
			if _, err := os.Stat(initFilePath); err == nil {
				return fmt.Errorf("%s already exists; use --force to overwrite", initFilePath)
			} else if !os.IsNotExist(err) {
				return err
			}
		}

		if err := os.WriteFile(initFilePath, []byte(defaultPipelineTemplate), 0644); err != nil {
			return err
		}

		fmt.Printf("Created %s\n", initFilePath)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initFilePath, "filepath", "f", "pin.yaml", "pipeline configuration file path")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing file")
	rootCmd.AddCommand(initCmd)
}
