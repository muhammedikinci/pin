package pin

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Pipeline struct {
	Workflow     []Job
	LogsWithTime bool
}

func parse() (Pipeline, error) {
	var pipeline Pipeline = Pipeline{}

	flows := viper.GetStringSlice("workflow")

	for _, v := range flows {
		configMap := viper.GetStringMap(v)

		job, err := generateJob(configMap)

		if err != nil {
			return Pipeline{}, err
		}

		job.Name = v

		pipeline.Workflow = append(pipeline.Workflow, job)
	}

	pipeline.LogsWithTime = viper.GetBool("logsWithTime")

	return pipeline, nil
}

func generateJob(configMap map[string]interface{}) (Job, error) {
	image, err := getJobImage(configMap["image"])

	if err != nil {
		return Job{}, err
	}

	script, err := getStringArray(configMap["script"])

	if err != nil {
		return Job{}, fmt.Errorf("`script` %w", err)
	}

	copyIgnore, err := getStringArray(configMap["copyignore"])

	if err != nil {
		return Job{}, fmt.Errorf("`copyIgnore` %w", err)
	}

	workDir, err := getWorkDir(configMap["workdir"])

	if err != nil {
		return Job{}, err
	}

	copyFiles, err := getCopyFiles(configMap["copyfiles"])

	if err != nil {
		return Job{}, err
	}

	soloExecution := getBool(configMap["soloexecution"], false)
	removeContainer := getBool(configMap["removecontainer"], true)
	port := getJobPort(configMap["port"])

	var job Job = Job{
		Image:           image,
		Script:          script,
		CopyFiles:       copyFiles,
		WorkDir:         workDir,
		SoloExecution:   soloExecution,
		RemoveContainer: removeContainer,
		Port:            port,
		CopyIgnore:      copyIgnore,
	}

	return job, nil
}

func getJobImage(image interface{}) (string, error) {
	if image == nil {
		return "", errors.New("image not specified")
	}

	return image.(string), nil
}

func getStringArray(stringArray interface{}) ([]string, error) {
	refVal := reflect.ValueOf(stringArray)

	if refVal.Kind() == reflect.Slice {
		arr := make([]string, refVal.Len())

		for i := 0; i < refVal.Len(); i++ {
			arr[i] = refVal.Index(i).Interface().(string)
		}

		return arr, nil
	}

	if refVal.Kind() == reflect.String {
		return []string{stringArray.(string)}, nil
	}

	return nil, errors.New("field is not valid")
}

func getJobPort(port interface{}) []Port {
	refVal := reflect.ValueOf(port)

	if refVal.Kind() == reflect.Slice {
		arr := make([]Port, refVal.Len())

		for i := 0; i < refVal.Len(); i++ {
			line := refVal.Index(i).Interface().(string)
			ports := strings.Split(line, ":")
			arr[i] = Port{Out: ports[0], In: ports[1]}
		}

		return arr
	}

	if refVal.Kind() == reflect.String {
		line := port.(string)
		ports := strings.Split(line, ":")
		return []Port{{Out: ports[0], In: ports[1]}}
	}

	return []Port{}
}

func getWorkDir(workDir interface{}) (string, error) {
	if workDir == nil {
		return "/root", nil
	}

	return workDir.(string), nil
}

func getCopyFiles(copyFiles interface{}) (bool, error) {
	if copyFiles == nil {
		return false, nil
	}

	return copyFiles.(bool), nil
}

func getBool(val interface{}, defaultValue bool) bool {
	if val == nil {
		return defaultValue
	}

	return val.(bool)
}
