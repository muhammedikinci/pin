package pin

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Apply(name string, filepath string) error {
	if err := checkFileExists(filepath); err != nil {
		return err
	}

	if err := readConfig(filepath); err != nil {
		return err
	}

	pipeline, err := parse()

	if err != nil {
		fmt.Println(err)
		return err
	}

	currentRunner := Runner{}

	if err := currentRunner.run(pipeline); err != nil {
		currentRunner.infoLog.Println(err.Error())
		return err
	}

	return nil
}

func checkFileExists(filepath string) error {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func readConfig(filepath string) error {
	fileBytes, err := os.ReadFile(filepath)

	if err != nil {
		return err
	}

	viper.SetConfigType("yaml")

	err = viper.ReadConfig(bytes.NewBuffer(fileBytes))

	if err != nil {
		return err
	}

	return nil
}
