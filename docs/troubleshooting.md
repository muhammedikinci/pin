# 🔧 Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using Pin.

## 🚨 Common Issues and Solutions

### Pipeline Validation Errors

#### Error: "either 'image' or 'dockerfile' must be specified"

**Problem**: Job configuration is missing both `image` and `dockerfile` fields.

```yaml
# ❌ Incorrect
job:
  script:
    - echo "hello"
```

**Solution**: Specify either an image or dockerfile:

```yaml
# ✅ Correct - Using image
job:
  image: alpine:latest
  script:
    - echo "hello"

# ✅ Correct - Using dockerfile
job:
  dockerfile: "./Dockerfile"
  script:
    - echo "hello"
```

#### Error: "port configuration validation failed"

**Problem**: Invalid port format in configuration.

```yaml
# ❌ Incorrect
port: "invalid:port:format:here"
```

**Solution**: Use valid port formats:

```yaml
# ✅ Correct formats
port:
  - "8080:80"                    # hostPort:containerPort
  - "127.0.0.1:8080:80"          # hostIP:hostPort:containerPort
  - "192.168.1.100:8080:80"      # specificIP:hostPort:containerPort
```

### Docker Connection Issues

#### Error: "Cannot connect to the Docker daemon"

**Problem**: Pin cannot connect to Docker daemon.

**Diagnostic Steps**:

```bash
# Check if Docker is running
docker ps

# Check Docker daemon status
systemctl status docker  # Linux
brew services list | grep docker  # macOS

# Check Docker socket permissions
ls -la /var/run/docker.sock
```

**Solutions**:

1. **Start Docker service**:
```bash
# Linux
sudo systemctl start docker

# macOS
open -a Docker
```

2. **Fix permissions** (Linux):
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

3. **Custom Docker host**:
```yaml
docker:
  host: "tcp://localhost:2375"  # For Docker Desktop
```

#### Error: "permission denied while trying to connect to Docker daemon"

**Problem**: User doesn't have permission to access Docker.

**Solution**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo (not recommended for regular use)
sudo pin apply -f pipeline.yaml
```

### Container Runtime Issues

#### Error: "container exited with non-zero status"

**Problem**: Script commands failed inside container.

**Diagnostic Steps**:

1. **Enable verbose logging**:
```yaml
logsWithTime: true
```

2. **Check individual commands**:
```yaml
# ❌ Commands run together (harder to debug)
script:
  - cd /app
  - npm install
  - npm test

# ✅ Commands run separately (easier to debug)
soloExecution: true
script:
  - cd /app
  - npm install
  - npm test
```

3. **Add debugging output**:
```yaml
script:
  - echo "Starting npm install..."
  - npm install
  - echo "npm install completed"
  - echo "Starting tests..."
  - npm test
```

#### Error: "no such file or directory"

**Problem**: File not found in container.

**Diagnostic Steps**:

1. **Check if files are copied**:
```yaml
# Enable file copying
copyFiles: true
```

2. **List container contents**:
```yaml
script:
  - ls -la
  - pwd
  - find . -name "package.json"
```

3. **Check copyIgnore settings**:
```yaml
copyIgnore:
  - "node_modules"  # Don't accidentally ignore needed files
  - ".git"
```

### Network and Port Issues

#### Error: "port already in use"

**Problem**: Port is occupied by another process.

**Diagnostic Steps**:
```bash
# Check what's using the port
lsof -i :8080
netstat -tulpn | grep 8080

# Kill process using port
kill -9 <PID>
```

**Solutions**:

1. **Use different port**:
```yaml
port:
  - "8081:8080"  # Use 8081 instead of 8080
```

2. **Bind to localhost only**:
```yaml
port:
  - "127.0.0.1:8080:8080"
```

#### Error: "connection refused"

**Problem**: Service not accessible on specified port.

**Diagnostic Steps**:

1. **Check if service is running**:
```yaml
script:
  - sleep 5  # Wait for service to start
  - curl -v http://localhost:8080/health
```

2. **Check port binding**:
```yaml
# Make sure service binds to 0.0.0.0, not just 127.0.0.1
script:
  - node server.js --host=0.0.0.0 --port=8080
```

### File System Issues

#### Error: "permission denied" when copying files

**Problem**: Container user doesn't have permission to access files.

**Solutions**:

1. **Use appropriate base image**:
```dockerfile
FROM node:18-alpine
# Some images run as non-root by default
USER root  # If needed
```

2. **Set proper file permissions**:
```yaml
script:
  - chmod +x ./scripts/deploy.sh
  - ./scripts/deploy.sh
```

### Environment Variable Issues

#### Error: "environment variable not found"

**Problem**: Environment variables not properly set.

**Diagnostic Steps**:

1. **List all environment variables**:
```yaml
script:
  - env | sort
  - echo "MY_VAR value: $MY_VAR"
```

2. **Check variable syntax**:
```yaml
# ✅ Correct
env:
  - MY_VAR=value
  - ANOTHER_VAR=another_value

# ❌ Incorrect
env:
  MY_VAR: value  # Wrong format
```

### Conditional Execution Issues

#### Error: "condition evaluation failed"

**Problem**: Invalid condition syntax.

**Solutions**:

1. **Check condition syntax**:
```yaml
# ✅ Correct
condition: $BRANCH == "main"
condition: $BRANCH == "main" && $DEPLOY == "true"
condition: $VAR != "test"

# ❌ Incorrect
condition: BRANCH == "main"  # Missing $
condition: $BRANCH = "main"  # Single = instead of ==
```

2. **Debug condition values**:
```yaml
script:
  - echo "BRANCH value: $BRANCH"
  - echo "DEPLOY value: $DEPLOY"
```

### Retry Mechanism Issues

#### Error: "retry configuration invalid"

**Problem**: Invalid retry parameters.

**Solutions**:

```yaml
# ✅ Correct retry configuration
retry:
  attempts: 3        # 1-10
  delay: 5          # 0-300 seconds
  backoff: 2.0      # 0.1-10.0

# ❌ Invalid configurations
retry:
  attempts: 15      # Too many attempts (max 10)
  delay: 500        # Too long delay (max 300)
  backoff: 0.05     # Too low backoff (min 0.1)
```

### Parallel Execution Issues

#### Error: "parallel jobs not running simultaneously"

**Problem**: Jobs marked as parallel but not executing in parallel.

**Solution**:

```yaml
# ✅ Correct parallel configuration
workflow:
  - job1
  - job2  # Both will run in parallel

job1:
  image: alpine:latest
  parallel: true
  script:
    - echo "Job 1 running"

job2:
  image: alpine:latest
  parallel: true
  script:
    - echo "Job 2 running"
```

## 🔍 Debugging Tools and Techniques

### Enable Verbose Logging

```yaml
# Add to your pipeline configuration
logsWithTime: true

# Use in scripts for debugging
script:
  - set -x  # Enable bash debug mode
  - echo "Debug: Starting command"
  - your-command
  - echo "Debug: Command completed"
```

### Container Debugging

```yaml
# Debug job to inspect container environment
debug:
  image: alpine:latest
  copyFiles: true
  script:
    - echo "=== System Information ==="
    - uname -a
    - echo "=== Environment Variables ==="
    - env | sort
    - echo "=== File System ==="
    - ls -la
    - df -h
    - echo "=== Network ==="
    - ip addr show || ifconfig
    - echo "=== Processes ==="
    - ps aux
```

### Network Debugging

```yaml
# Network connectivity debug job
network-debug:
  image: alpine:latest
  script:
    - apk add --no-cache curl netcat-openbsd
    - echo "=== Testing DNS ==="
    - nslookup google.com
    - echo "=== Testing HTTP ==="
    - curl -v http://google.com
    - echo "=== Testing Port Connectivity ==="
    - nc -zv google.com 80
```

## 🚀 Performance Troubleshooting

### Slow Pipeline Execution

**Diagnostic Steps**:

1. **Profile image pull times**:
```yaml
script:
  - echo "Image pull completed at $(date)"
```

2. **Optimize Docker images**:
```yaml
# Use specific, smaller images
image: node:18-alpine  # Instead of node:latest
image: golang:1.21-alpine  # Instead of golang:latest
```

3. **Minimize file copying**:
```yaml
copyFiles: true
copyIgnore:
  - "node_modules"
  - ".git"
  - "*.log"
  - "coverage"
```

### Memory Issues

**Diagnostic Steps**:

```yaml
script:
  - echo "=== Memory Usage ==="
  - free -h
  - echo "=== Disk Usage ==="
  - df -h
```

**Solutions**:

1. **Use lightweight base images**:
```yaml
image: alpine:latest  # ~5MB
# Instead of ubuntu:latest (~70MB)
```

2. **Clean up in scripts**:
```yaml
script:
  - npm install
  - npm run build
  - rm -rf node_modules  # Clean up after build
```

## 📞 Getting Help

### Collect Debug Information

When reporting issues, include:

1. **Pin version**:
```bash
pin --version
```

2. **Pipeline configuration**:
```yaml
# Your pipeline.yaml content
```

3. **Error output**:
```bash
pin apply -f pipeline.yaml 2>&1 | tee debug.log
```

4. **System information**:
```bash
# Operating system
uname -a

# Docker version
docker --version

# Available resources
docker system df
```

### Common Log Patterns

Look for these patterns in logs:

- `validation error`: Configuration issues
- `connection refused`: Network/port problems
- `permission denied`: File/Docker permissions
- `no such file`: Missing files or incorrect paths
- `exit status 1`: Script command failures

### Enable Debug Mode

```bash
# Run with maximum verbosity
pin apply -f pipeline.yaml --verbose

# Or set environment variable
DEBUG=* pin apply -f pipeline.yaml
```

This troubleshooting guide covers the most common issues users encounter with Pin. For additional help, check the [GitHub Issues](https://github.com/muhammedikinci/pin/issues) page or create a new issue with detailed information about your problem.