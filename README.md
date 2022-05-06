# pin 🔥

WIP - Local pipeline command line and web interface project with Docker Golang API.

![pingif](https://user-images.githubusercontent.com/11901620/166977370-9526a377-41d6-4c96-a2b3-56f57ee4edd1.gif)

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

run:
  image: golang:alpine3.15
  copyFiles: true
  soloExecution: true
  script:
    - go mod download
    - go run .
    - ls
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

# Checklist

- Implement web interface
- Support concurrent jobs
- Add working with remote docker deamon support
- Change image pulling logs (get only status logs)✅[Issue#1](https://github.com/muhammedikinci/pin/issues/1)
- Add custom ignore configuration to copyFiles for project files (like gitignore)
- Add shared artifacts support between different jobs 

# Contact

Muhammed İkinci - muhammedikinci@outlook.com
