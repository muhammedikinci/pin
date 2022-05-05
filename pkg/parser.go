package pin

import (
	"errors"
	"reflect"

	"github.com/spf13/viper"
)

type Workflow []Job

func parse() (Workflow, error) {
	var workflow Workflow = Workflow{}

	flows := viper.GetStringSlice("workflow")

	for _, v := range flows {
		configMap := viper.GetStringMap(v)

		job, err := generateJob(configMap)

		if err != nil {
			return nil, err
		}

		job.Name = v

		workflow = append(workflow, job)
	}

	return workflow, nil
}

func generateJob(configMap map[string]interface{}) (Job, error) {
	image, err := getJobImage(configMap["image"])

	if err != nil {
		return Job{}, err
	}

	script, err := getJobScripts(configMap["script"])

	if err != nil {
		return Job{}, err
	}

	workDir, err := getWorkDir(configMap["workdir"])

	if err != nil {
		return Job{}, err
	}

	copyFiles, err := getCopyFiles(configMap["copyfiles"])

	if err != nil {
		return Job{}, err
	}

	soloExecution := getBool(configMap["soloexecution"])

	if err != nil {
		return Job{}, err
	}

	var job Job = Job{
		Image:         image,
		Script:        script,
		CopyFiles:     copyFiles,
		WorkDir:       workDir,
		SoloExecution: soloExecution,
	}

	return job, nil
}

func getJobImage(image interface{}) (string, error) {
	if image == nil {
		return "", errors.New("image not specified")
	}

	return image.(string), nil
}

func getJobScripts(script interface{}) ([]string, error) {
	refVal := reflect.ValueOf(script)

	if refVal.Kind() == reflect.Slice {
		arr := make([]string, refVal.Len())

		for i := 0; i < refVal.Len(); i++ {
			arr[i] = refVal.Index(i).Interface().(string)
		}

		return arr, nil
	}

	if refVal.Kind() == reflect.String {
		return []string{script.(string)}, nil
	}

	return nil, errors.New("`script` field is not valid")
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

func getBool(val interface{}) bool {
	if val == nil {
		return false
	}

	return val.(bool)
}
