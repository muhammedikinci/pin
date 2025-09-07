package runner

import (
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Pipeline struct {
	Workflow     []*Job
	LogsWithTime bool
	DockerHost   string
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
	pipeline.DockerHost = viper.GetString("docker.host")

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
	dockerfile := getString(configMap["dockerfile"])
	retryConfig := getRetryConfig(configMap["retry"])

	var job *Job = &Job{
		Image:         image,
		Dockerfile:    dockerfile,
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
		RetryConfig:   retryConfig,
	}

	return job, nil
}

func getJobImage(image interface{}) (string, error) {
	if image == nil {
		return "", nil
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
			arr[i] = parsePortString(line)
		}

		return arr
	}

	if refVal.Kind() == reflect.String {
		line := port.(string)
		return []Port{parsePortString(line)}
	}

	return []Port{}
}

// parsePortString parses port configuration string into Port struct
// Supports formats:
// - "8080:80" -> hostIP: "0.0.0.0", hostPort: "8080", containerPort: "80"
// - "127.0.0.1:8080:80" -> hostIP: "127.0.0.1", hostPort: "8080", containerPort: "80"
// - "localhost:8080:80" -> hostIP: "localhost", hostPort: "8080", containerPort: "80"
func parsePortString(portStr string) Port {
	parts := strings.Split(portStr, ":")
	
	switch len(parts) {
	case 2:
		// Format: "8080:80"
		return Port{
			Out:    parts[0],
			In:     parts[1],
			HostIP: "0.0.0.0", // Default host IP
		}
	case 3:
		// Format: "127.0.0.1:8080:80" or "localhost:8080:80"
		return Port{
			HostIP: parts[0],
			Out:    parts[1],
			In:     parts[2],
		}
	default:
		// Fallback to default format if invalid
		return Port{
			Out:    "8080",
			In:     "80",
			HostIP: "0.0.0.0",
		}
	}
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

// getRetryConfig parses retry configuration from the config map
func getRetryConfig(retryInterface interface{}) RetryConfig {
	// Default retry config (no retry)
	defaultConfig := RetryConfig{
		MaxAttempts:       1,
		DelaySeconds:     1,
		BackoffMultiplier: 1.0,
	}

	if retryInterface == nil {
		return defaultConfig
	}

	retryMap, ok := retryInterface.(map[string]interface{})
	if !ok {
		return defaultConfig
	}

	config := defaultConfig

	if maxAttempts := retryMap["attempts"]; maxAttempts != nil {
		if attempts, ok := maxAttempts.(int); ok && attempts > 0 {
			config.MaxAttempts = attempts
		}
	}

	if delay := retryMap["delay"]; delay != nil {
		if delaySeconds, ok := delay.(int); ok && delaySeconds >= 0 {
			config.DelaySeconds = delaySeconds
		}
	}

	if backoff := retryMap["backoff"]; backoff != nil {
		if multiplier, ok := backoff.(float64); ok && multiplier > 0 {
			config.BackoffMultiplier = multiplier
		}
	}

	return config
}
