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

Sample yaml file

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

You can use this feature for port forwarding from container to your machine with multiple mapping

```yaml
port:
  - 8082:8080
  - 8083:8080
```

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

# Tests

```sh
go test ./...
```

# Contact

Muhammed İkinci - muhammedikinci@outlook.com

