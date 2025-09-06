<p align="center">
  <br>
    <img src="asset/pin.png" width="200"/>
  <br>
</p>

# pin 🔥 [![pipeline](https://github.com/muhammedikinci/pin/actions/workflows/go.yml/badge.svg)](https://github.com/muhammedikinci/pin/actions/workflows/go.yml)

WIP - Local pipeline project with Docker Golang API.

![pingif](asset/pin.gif)

<sup><sup>terminal from [terminalgif.com](https://terminalgif.com)</sup></sup>

# 🌐 Installation

## Download latest release

You can download latest release from [here](https://github.com/muhammedikinci/pin/releases)

## Install with cloning

Clone the pin

```sh
git clone https://github.com/muhammedikinci/pin
```

Download packages

```sh
go mod download
```

Build executable

```sh
go build -o pin ./cmd/cli/.
```

Or you can run directly

```sh
go run ./cmd/cli/. apply -n "test" -f ./testdata/test.yaml
```

# ⚙️ Configuration

Pin includes built-in YAML validation to catch configuration errors before pipeline execution.

## Pipeline Validation

Pin automatically validates your pipeline configuration before execution:

- ✅ **Required fields**: Ensures either `image` or `dockerfile` is specified
- ✅ **Field types**: Validates all fields have correct data types  
- ✅ **Port formats**: Checks port configurations match supported formats
- ✅ **Script validation**: Ensures scripts are not empty
- ✅ **Boolean fields**: Validates boolean configurations

### Validation Examples

```bash
# Valid configuration passes validation
$ pin apply -f pipeline.yaml
Pipeline validation successful
⚉ build Starting...

# Invalid configuration shows helpful errors
$ pin apply -f invalid.yaml
Pipeline validation failed: validation error in job 'build': either 'image' or 'dockerfile' must be specified
```

## Sample yaml file

```yaml
workflow:
  - run

logsWithTime: true

run:
  image: golang:alpine3.15
  copyFiles: true
  soloExecution: true
  script:
    - go mod download
    - go run .
    - ls
  port:
    - 8082:8080
```

You can create separate jobs like the `run` stage and if you want to run these jobs in the pipeline you must add its name to `workflow`.

## copyFiles

default: false

If you want to copy all projects filed to the docker container, you must set this configuration to `true`

## soloExecution

default: false

When you add multiple commands to the `script` field, commands are running in the container as a shell script. If soloExecution is set to `true` each command works in a different shell script.

#### soloExecution => false

```sh
# shell#1
cd cmd
ls
```

#### soloExecution => true

```sh
# shell#1
cd cmd
```

```sh
# shell#2
ls
```

If you want to see all files in the cmd folder you must set soloExecution to false or you can use this:

```sh
# shell#1
cd cmd && ls
```

## logsWithTime

default: false

logsWithTime => true

```sh
⚉ 2022/05/08 11:36:30 Image is available
⚉ 2022/05/08 11:36:30 Start creating container
⚉ 2022/05/08 11:36:33 Starting the container
⚉ 2022/05/08 11:36:35 Execute command: ls -a
```

logsWithTime => false

```sh
⚉ Image is available
⚉ Start creating container
⚉ Starting the container
⚉ Execute command: ls -a
```

## port

default: empty mapping

You can use this feature for port forwarding from container to your machine with flexible host and port configuration.

### Port Configuration Formats

1. **Standard format**: `"hostPort:containerPort"`
2. **Custom host format**: `"hostIP:hostPort:containerPort"`

### Examples

```yaml
# Standard port mapping (binds to all interfaces)
port: "8080:80"

# Multiple ports with different configurations
port:
  - "8082:8080"                    # Standard format
  - "127.0.0.1:8083:8080"          # Bind only to localhost
  - "192.168.1.100:8084:8080"      # Bind to specific IP address

# Mix of standard and custom host formats
run:
  image: nginx:alpine
  port:
    - "8080:80"                    # Available on all network interfaces
    - "127.0.0.1:8081:80"          # Only accessible from localhost
    - "0.0.0.0:8082:80"            # Explicitly bind to all interfaces
```

### Use Cases

- **Security**: Bind services only to localhost (`127.0.0.1:8080:80`)
- **Network isolation**: Bind to specific network interfaces (`192.168.1.100:8080:80`)
- **Development**: Expose different ports for different environments

## copyIgnore

default: empty mapping

You can use this feature to ignore copying the specific files in your project to the container.

Sample configuration yaml

```yaml
run:
  image: node:current-alpine3.15
  copyFiles: true
  soloExecution: true
  port:
    - 8080:8080
  copyIgnore:
    - server.js
    - props
    - README.md
    - helper/.*/.py
```

Actual folder structure in project

```yaml
index.js
server.js
README.md
helper:
    - test.py
    - mock
        test2.py
    - api:
        index.js
    - props:
        index.js
```

Folder structure in container

```yaml
index.js
helper:
    - mock (empty)
    - api:
        index.js
```

## parallel

default: false

If you want to run parallel job, you must add `parallel` field and the stage must be in workflow(position doesn't matter)

```yaml
workflow:
  - testStage
  - parallelJob
  - run
---
parallelJob:
  image: node:current-alpine3.15
  copyFiles: true
  soloExecution: true
  parallel: true
  script:
    - ls -a
```

## Environment Variables

You can specify environment variables for your jobs in the YAML configuration. These variables will be available inside the container during job execution.

Example:

```yaml
workflow:
  - run

run:
  image: golang:alpine3.15
  copyFiles: true
  soloExecution: true
  script:
    - go mod download
    - go run .
    - echo "Environment variables:"
    - echo "MY_VAR: $MY_VAR"
    - echo "ANOTHER_VAR: $ANOTHER_VAR"
  port:
    - 8082:8080
  env:
    - MY_VAR=value
    - ANOTHER_VAR=another_value
```

In this example, the environment variables `MY_VAR` and `ANOTHER_VAR` are set and printed during job execution.

## Conditional Execution

You can specify conditions for job execution using the `condition` field. Jobs will only run if the condition evaluates to true.

Example:

```yaml
workflow:
  - build
  - test
  - deploy

build:
  image: golang:alpine3.15
  copyFiles: true
  script:
    - go build -o app .

test:
  image: golang:alpine3.15
  copyFiles: true
  script:
    - go test ./...

deploy:
  image: alpine:latest
  condition: $BRANCH == "main"
  script:
    - echo "Deploying to production..."
    - ./deploy.sh
```

### Supported Condition Operators

- **Equality**: `$VAR == "value"` - Check if variable equals value
- **Inequality**: `$VAR != "value"` - Check if variable does not equal value
- **AND**: `$VAR1 == "value1" && $VAR2 == "value2"` - Both conditions must be true
- **OR**: `$VAR1 == "value1" || $VAR2 == "value2"` - At least one condition must be true
- **Variable existence**: `$VAR` - Check if variable exists and is not empty/false/0

### Examples

```yaml
# Run only on main branch
deploy:
  condition: $BRANCH == "main"

# Run on main or develop branch
deploy:
  condition: $BRANCH == "main" || $BRANCH == "develop"

# Run only when both conditions are met
deploy:
  condition: $BRANCH == "main" && $DEPLOY == "true"

# Run when variable exists
cleanup:
  condition: $CLEANUP_ENABLED

# Run when environment is not test
deploy:
  condition: $ENV != "test"
```

You can set environment variables before running pin:

```bash
BRANCH=main pin apply -f pipeline.yaml
```

## Custom Dockerfile

You can use a custom Dockerfile to build your own image for the job instead of pulling a pre-built image.

Example:

```yaml
workflow:
  - custom-build

custom-build:
  dockerfile: "./Dockerfile"
  copyFiles: true
  script:
    - echo "Hello from custom Docker image!"
    - ls -la
```

### Key Features

- **dockerfile**: Path to your custom Dockerfile
- **Automatic image building**: Pin will build the image from your Dockerfile before running the job
- **Build context**: The directory containing the Dockerfile will be used as the build context
- **Image naming**: Built images are automatically tagged as `<job-name>-custom:latest`

### Example Dockerfile

```dockerfile
FROM alpine:latest

RUN apk add --no-cache \
    bash \
    curl \
    git \
    make

WORKDIR /app
USER nobody

CMD ["/bin/bash"]
```

**Note**: When using `dockerfile`, you don't need to specify the `image` field. Pin will use the built image automatically.

# Tests

```sh
go test ./...
```

# Contact

Muhammed İkinci - muhammedikinci@outlook.com

