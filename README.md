<p align="center">
  <br>
    <img src="asset/pin.png" width="200"/>
  <br>
</p>

# pin 🔥 [![pipeline](https://github.com/muhammedikinci/pin/actions/workflows/go.yml/badge.svg)](https://github.com/muhammedikinci/pin/actions/workflows/go.yml)

Pin is a tiny pipeline runner for developers who want to run build, test, and deploy jobs on their own VPS without setting up a full CI system.

![pingif](asset/pin.gif)

<sup><sup>terminal from [terminalgif.com](https://terminalgif.com)</sup></sup>

## Why Pin?

Use Pin when you have Docker on a machine and want a simple way to:

- run the same pipeline locally or on a VPS
- trigger deploys over HTTP
- watch live pipeline events with Server-Sent Events
- avoid maintaining a full CI server for small projects

Pin is not trying to replace GitHub Actions, Jenkins, or Kubernetes. It is for the smaller, very common case: "I have a server, I have Docker, I want to build/test/deploy safely from one small binary."

## Quick Start

Create a starter pipeline:

```bash
pin init
pin validate -f pin.yaml
pin run -f pin.yaml
```

Run Pin as a daemon on your VPS:

```bash
PIN_TOKEN=change-me pin daemon --host 0.0.0.0 --port 8081

# From your laptop or another machine
pin trigger -f pin.yaml --url http://your-vps:8081 --token change-me
pin watch --url http://your-vps:8081 --token change-me
pin runs --url http://your-vps:8081 --token change-me
pin logs <run_id> --url http://your-vps:8081 --token change-me
```

Check your local setup:

```bash
pin doctor -f pin.yaml --url http://your-vps:8081 --token change-me
```

## Core Commands

| Command | Description |
| --- | --- |
| `pin init` | Create a starter `pin.yaml` |
| `pin validate -f pin.yaml` | Validate config without running Docker jobs |
| `pin run -f pin.yaml` | Run a pipeline locally |
| `pin daemon --host 127.0.0.1 --port 8081` | Start the HTTP/SSE daemon |
| `pin trigger -f pin.yaml --url http://server:8081` | Send a pipeline to a daemon |
| `pin watch --url http://server:8081` | Watch live daemon events |
| `pin runs --url http://server:8081` | List recent daemon runs |
| `pin logs <run_id> --url http://server:8081` | Fetch persisted logs for one run |
| `pin doctor` | Check Docker, config, port, and daemon access |

The older `pin apply -f pin.yaml` command still works as a backward-compatible alias for local execution.

## Daemon Mode

Pin can run as a long-running daemon with an HTTP API and Server-Sent Events stream. This is the main VPS workflow: keep one `pin daemon` process running, trigger pipelines remotely, and watch them live.

Runs and logs are persisted under `.pin/runs` by default. Use `--data-dir /var/lib/pin` for production daemons.

### HTTP Endpoints

| Endpoint   | Method | Description                                     |
| ---------- | ------ | ----------------------------------------------- |
| `/events`  | GET    | Server-Sent Events stream for real-time updates |
| `/health`  | GET    | Health check and connected client count         |
| `/trigger` | POST   | Trigger pipeline execution with YAML config     |
| `/runs`    | GET    | Recent pipeline run statuses                    |
| `/runs/{id}` | GET | Details for one run |
| `/runs/{id}/logs` | GET | Persisted log output for one run |
| `/webhooks/github` | POST | GitHub push webhook endpoint |
| `/`        | GET    | API information and available endpoints         |

### Daemon Security

The daemon binds to `127.0.0.1` by default. If you expose it with `--host 0.0.0.0`, set a token:

```bash
PIN_TOKEN=change-me pin daemon --host 0.0.0.0 --port 8081
```

Clients can pass the token with `--token`, the `PIN_TOKEN` environment variable, `Authorization: Bearer <token>`, `X-Pin-Token`, or a `?token=` query parameter for EventSource-style clients.

### GitHub Webhooks

Pin can trigger a configured pipeline from GitHub push events:

```bash
PIN_TOKEN=change-me PIN_GITHUB_SECRET=github-secret \
  pin daemon \
  --host 127.0.0.1 \
  --port 8081 \
  --data-dir /var/lib/pin \
  --github-pipeline /opt/myapp/pin.yaml \
  --github-branch main
```

Set the GitHub webhook payload URL to `https://your-domain/webhooks/github`, content type to `application/json`, and secret to the same value as `PIN_GITHUB_SECRET`.

### Host Jobs

Most jobs run in Docker containers with `image` or `dockerfile`. Deploy steps that need direct VPS access can use `host: true`:

```yaml
deploy:
  host: true
  condition: $BRANCH == "main"
  script:
    - docker compose up -d --build
```

Host jobs run as the daemon user, so keep that user scoped to the permissions your deploy needs.

### Real-time Events

The daemon broadcasts various events during pipeline execution:

- **daemon_start**: Service started successfully
- **pipeline_trigger**: New pipeline execution requested
- **pipeline_started**: Queued pipeline began running
- **job_container_start**: Container started for job
- **log**: Real-time log messages from jobs
- **job_completed**: Job finished successfully
- **job_failed**: Job failed with error details
- **pipeline_complete**: Entire pipeline finished
- **daemon_stop**: Service shutting down

### Example Event Stream

```javascript
// Connect to event stream
const eventSource = new EventSource("http://localhost:8081/events?token=change-me");

eventSource.onmessage = function (event) {
  const data = JSON.parse(event.data);
  console.log(`[${data.level}] ${data.message}`);
};

// Events received:
// {"level":"info","message":"Pipeline execution started","job":"build"}
// {"level":"info","message":"Container started","job":"build"}
// {"level":"success","message":"Job completed successfully","job":"build"}
```

# 🌐 Installation

## Download latest release

You can download latest release from [here](https://github.com/muhammedikinci/pin/releases)

Or install the latest release with:

```sh
curl -fsSL https://raw.githubusercontent.com/muhammedikinci/pin/main/scripts/install.sh | sh
```

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
go run ./cmd/cli/. run -f ./testdata/test.yaml
```

# ⚙️ Configuration

Pin includes built-in YAML validation to catch configuration errors before pipeline execution.

## Pipeline Validation

Pin automatically validates your pipeline configuration before execution:

- ✅ **Required fields**: Ensures either `image`, `dockerfile`, or `host: true` is specified
- ✅ **Field types**: Validates all fields have correct data types
- ✅ **Port formats**: Checks port configurations match supported formats
- ✅ **Script validation**: Ensures scripts are not empty
- ✅ **Boolean fields**: Validates boolean configurations

### Validation Examples

```bash
# Valid configuration passes validation
$ pin validate -f pipeline.yaml
Pipeline configuration is valid: pipeline.yaml

# Invalid configuration shows helpful errors
$ pin validate -f invalid.yaml
Pipeline validation failed: validation error in job 'build': either 'image' or 'dockerfile' must be specified
```

## Sample yaml file

```yaml
workflow:
  - run

logsWithTime: true

# Optional: Specify custom Docker host
docker:
  host: "tcp://localhost:2375"

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

## docker

Configure Docker daemon connection settings.

### host

default: system default (usually `unix:///var/run/docker.sock` on Linux/macOS)

Specify a custom Docker host to connect to a different Docker daemon. This is useful for:

- **Remote Docker**: Connect to Docker running on another machine
- **Docker Desktop**: Connect to Docker Desktop on different ports
- **CI/CD environments**: Connect to specific Docker instances
- **Development**: Switch between local and remote Docker instances

### Supported Docker Host Formats

```yaml
# TCP connection to remote Docker daemon
docker:
  host: "tcp://192.168.1.100:2375"

# TCP connection with TLS (secure)
docker:
  host: "tcp://docker.example.com:2376"

# Unix socket (Linux/macOS default)
docker:
  host: "unix:///var/run/docker.sock"

# Windows named pipe
docker:
  host: "npipe://./pipe/docker_engine"

# SSH connection to remote host
docker:
  host: "ssh://user@docker-host"
```

### Examples

```yaml
# Connect to local Docker Desktop
workflow:
  - build

docker:
  host: "tcp://localhost:2375"

build:
  image: golang:alpine
  script:
    - go build .

# Connect to remote Docker daemon
workflow:
  - deploy

docker:
  host: "tcp://production-docker:2375"

deploy:
  image: alpine:latest
  script:
    - echo "Deploying to remote Docker"
```

### Security Notes

- Use TLS (port 2376) for remote connections in production
- Ensure Docker daemon is properly secured when exposing TCP ports
- Consider using SSH tunneling for secure remote connections

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

## Job Retry Mechanism

Pin supports automatic job retries with configurable parameters for handling transient failures.

### retry

default: no retry (attempts: 1)

Configure automatic retry behavior for jobs that fail due to temporary issues like network problems, resource constraints, or external service unavailability.

### Retry Configuration Options

```yaml
retry:
  attempts: 3 # Number of attempts (1-10, default: 1)
  delay: 5 # Initial delay in seconds (0-300, default: 1)
  backoff: 2.0 # Exponential backoff multiplier (0.1-10.0, default: 1.0)
```

### Examples

```yaml
# Simple retry - 3 attempts with 2 second delays
workflow:
  - unstable-service

unstable-service:
  image: alpine:latest
  retry:
    attempts: 3
    delay: 2
  script:
    - echo "Attempting to connect to service..."
    - curl https://unstable-api.example.com/health

# Advanced retry with exponential backoff
workflow:
  - network-dependent

network-dependent:
  image: alpine:latest
  retry:
    attempts: 5      # Try 5 times total
    delay: 1         # Start with 1 second delay
    backoff: 2.0     # Double delay each retry (1s, 2s, 4s, 8s)
  script:
    - wget https://external-resource.com/data.zip
```

### Retry Behavior

1. **Linear Delays**: With `backoff: 1.0`, delays remain constant
2. **Exponential Backoff**: With `backoff > 1.0`, delays increase exponentially
3. **Failure Logging**: Each retry attempt is logged with reason and next attempt time
4. **Final Failure**: After all attempts fail, the job fails with the last error

### Use Cases

- **Network Operations**: Downloads, API calls, external service connections
- **Resource Competition**: Database connections, file locks, temporary resource unavailability
- **CI/CD Pipelines**: Flaky tests, temporary infrastructure issues
- **External Dependencies**: Third-party services, cloud resources

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
BRANCH=main pin run -f pipeline.yaml
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

## 📚 Documentation

For comprehensive documentation, examples, and guides:

- **[📖 Complete Documentation](docs/README.md)** - Full documentation index
- **[🚀 Examples](docs/examples.md)** - Practical examples and use cases
- **[🌐 API Reference](docs/api-reference.md)** - HTTP API documentation for daemon mode
- **[🔧 Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions
- **[🎯 Use Cases](docs/use-cases.md)** - Real-world applications and workflows

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📞 Support

- [GitHub Issues](https://github.com/muhammedikinci/pin/issues) - Bug reports and feature requests
- [GitHub Discussions](https://github.com/muhammedikinci/pin/discussions) - Community discussions

# Contact

Muhammed İkinci - muhammedikinci@outlook.com
