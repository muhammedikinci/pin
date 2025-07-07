package runner

import (
	"errors"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Pipeline struct {
	Workflow     []*Job
	LogsWithTime bool
}

func parse() (Pipeline, error) {
	var pipeline Pipeline = Pipeline{}

	flows := viper.GetStringSlice("workflow")

	for i, v := range flows {
		configMap := viper.GetStringMap(v)

		job, err := generateJob(configMap)
		if err != nil {
			return Pipeline{}, err
		}

		job.Name = v

		if i > 0 && (!job.IsParallel || !pipeline.Workflow[i-1].IsParallel) {
			job.Previous = pipeline.Workflow[i-1]
		}

		pipeline.Workflow = append(pipeline.Workflow, job)
	}

	pipeline.LogsWithTime = viper.GetBool("logsWithTime")

	return pipeline, nil
}

func generateJob(configMap map[string]interface{}) (*Job, error) {
	image, err := getJobImage(configMap["image"])
	if err != nil {
		return &Job{}, err
	}

	workDir, err := getWorkDir(configMap["workdir"])
	if err != nil {
		return &Job{}, err
	}

	copyFiles, err := getCopyFiles(configMap["copyfiles"])
	if err != nil {
		return &Job{}, err
	}

	soloExecution := getBool(configMap["soloexecution"], false)
	isParallel := getBool(configMap["parallel"], false)
	copyIgnore := getStringArray(configMap["copyignore"])
	script := getStringArray(configMap["script"])
	port := getJobPort(configMap["port"])
	env := getEnv(configMap["env"])
	artifactPath := getString(configMap["artifactpath"])
	condition := getString(configMap["condition"])

	var job *Job = &Job{
		Image:         image,
		Script:        script,
		CopyFiles:     copyFiles,
		WorkDir:       workDir,
		SoloExecution: soloExecution,
		IsParallel:    isParallel,
		Port:          port,
		CopyIgnore:    copyIgnore,
		ErrorChannel:  make(chan error, 1),
		Env:           env,
		ArtifactPath:  artifactPath,
		Condition:     condition,
	}

	return job, nil
}

func getJobImage(image interface{}) (string, error) {
	if image == nil {
		return "", errors.New("image not specified")
	}

	return image.(string), nil
}

func getStringArray(stringArray interface{}) []string {
	refVal := reflect.ValueOf(stringArray)

	if refVal.Kind() == reflect.Slice {
		arr := make([]string, refVal.Len())

		for i := 0; i < refVal.Len(); i++ {
			arr[i] = refVal.Index(i).Interface().(string)
		}

		return arr
	}

	if refVal.Kind() == reflect.String {
		return []string{stringArray.(string)}
	}

	return []string{}
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

func getEnv(env interface{}) []string {
	return getStringArray(env)
}

func getString(val interface{}) string {
	if val == nil {
		return ""
	}
	return val.(string)
}
