<p align="center">
  <br>
    <img src="asset/pin.png" width="200"/>
  <br>
</p>

# pin 🔥 [<!--lint ignore no-dead-urls-->![GitHub Actions status | sdras/awesome-actions](https://github.com/muhammedikinci/pin/actions/workflows/go.yml/badge.svg)](https://github.com/muhammedikinci/pin/actions/workflows/go.yml)

WIP - Local pipeline command line and web interface project with Docker Golang API.

![pingif](asset/pin.gif)

# 🌐 Installation 

Clone the pin

```sh
git clone https://github.com/muhammedikinci/pin
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

You can create separate jobs like the `run` stage and if you want to run these jobs in the pipeline you must add its name to `workflow`. Jobs only work serialized for now.

## copyFiles

default: false

If you want to copy all projects filed to the docker container, you must set this configuration to `true`

custom folder ignore doesn't support yet!

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

# Tests

```sh
go test ./...
```

# Checklist

- Implement web interface
- Support concurrent jobs
- Add working with remote docker deamon support
- Change image pulling logs (get only status logs)✅[Issue#1](https://github.com/muhammedikinci/pin/issues/1)
- Add custom ignore configuration to copyFiles for project files (like gitignore) ✅[Issue#7](https://github.com/muhammedikinci/pin/issues/7)
- Add shared artifacts support between different jobs 
- Add timestamp to container names ✅[Issue#2](https://github.com/muhammedikinci/pin/issues/2)
- Create small pieces with extracting codes from runner struct and write unit test:
  - Image Manager ✅[Issue#3](https://github.com/muhammedikinci/pin/issues/3)
  - Container Manager ✅[Issue#4](https://github.com/muhammedikinci/pin/issues/4)
  - Shell Commander ✅[Issue#5](https://github.com/muhammedikinci/pin/issues/5)
  - Parser
  - Runner
- Add port expose support ✅[Issue#6](https://github.com/muhammedikinci/pin/issues/6)
- Support long living containers
- Add concurrency between jobs
- Add graceful shutdown with context 

# Contact

Muhammed İkinci - muhammedikinci@outlook.com
