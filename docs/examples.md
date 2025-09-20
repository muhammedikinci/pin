# 📚 Pin Examples

This document contains practical examples for different use cases of Pin.

## 🏗️ Basic Examples

### 1. Building a Simple Go Project

```yaml
# build.yaml
workflow:
  - build
  - test

build:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go mod download
    - go build -o app ./cmd/main.go
    - ls -la

test:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test ./...
```

```bash
pin apply -f build.yaml
```

### 2. Node.js Application Development

```yaml
# node-dev.yaml
workflow:
  - install
  - lint
  - test
  - build

logsWithTime: true

install:
  image: node:18-alpine
  copyFiles: true
  script:
    - npm ci

lint:
  image: node:18-alpine
  copyFiles: true
  script:
    - npm run lint

test:
  image: node:18-alpine
  copyFiles: true
  script:
    - npm test

build:
  image: node:18-alpine
  copyFiles: true
  script:
    - npm run build
    - ls -la dist/
```

### 3. Multi-Stage Docker Build

```yaml
# docker-build.yaml
workflow:
  - build-image
  - test-image
  - security-scan

build-image:
  dockerfile: "./Dockerfile"
  copyFiles: true
  script:
    - echo "Building custom image..."

test-image:
  image: docker:latest
  script:
    - docker run --rm build-image-custom:latest echo "Image test successful"

security-scan:
  image: aquasec/trivy:latest
  script:
    - trivy image build-image-custom:latest
```

## 🌐 Web Application Examples

### 4. React Application Development

```yaml
# react-dev.yaml
workflow:
  - install
  - start

logsWithTime: true

install:
  image: node:18-alpine
  copyFiles: true
  script:
    - npm install

start:
  image: node:18-alpine
  copyFiles: true
  port:
    - "3000:3000"
  script:
    - npm start
```

### 5. Full-Stack Application (Backend + Frontend)

```yaml
# fullstack.yaml
workflow:
  - backend
  - frontend

backend:
  image: node:18-alpine
  copyFiles: true
  port:
    - "127.0.0.1:8000:8000"
  parallel: true
  script:
    - cd backend
    - npm install
    - npm start

frontend:
  image: node:18-alpine
  copyFiles: true
  port:
    - "3000:3000"
  parallel: true
  script:
    - cd frontend
    - npm install
    - npm start
```

## 🧪 Testing Scenarios

### 6. Unit and Integration Tests

```yaml
# testing.yaml
workflow:
  - unit-tests
  - integration-tests
  - coverage-report

unit-tests:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test -short ./...

integration-tests:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test -tags=integration ./...

coverage-report:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test -coverprofile=coverage.out ./...
    - go tool cover -html=coverage.out -o coverage.html
    - echo "Coverage report generated: coverage.html"
```

### 7. Database Testing

```yaml
# database-test.yaml
workflow:
  - start-db
  - run-migrations
  - run-tests
  - cleanup

start-db:
  image: postgres:15-alpine
  port:
    - "5432:5432"
  env:
    - POSTGRES_DB=testdb
    - POSTGRES_USER=testuser
    - POSTGRES_PASSWORD=testpass
  script:
    - sleep 5  # Wait for DB to be ready

run-migrations:
  image: migrate/migrate:latest
  script:
    - migrate -path ./migrations -database "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable" up

run-tests:
  image: golang:1.21-alpine
  copyFiles: true
  env:
    - DATABASE_URL=postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable
  script:
    - go test ./...

cleanup:
  image: postgres:15-alpine
  script:
    - echo "Database tests completed"
```

## 🚀 Deployment Examples

### 8. Branch-Based Deployment

```yaml
# deployment.yaml
workflow:
  - build
  - test
  - deploy-staging
  - deploy-production

build:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go build -o app .

test:
  image: golang:1.21-alpine
  copyFiles: true
  script:
    - go test ./...

deploy-staging:
  image: alpine:latest
  condition: $BRANCH == "develop"
  script:
    - echo "Deploying to staging environment..."
    - echo "Staging deployment completed"

deploy-production:
  image: alpine:latest
  condition: $BRANCH == "main"
  script:
    - echo "Deploying to production environment..."
    - echo "Production deployment completed"
```

Usage:
```bash
# Staging deployment
BRANCH=develop pin apply -f deployment.yaml

# Production deployment
BRANCH=main pin apply -f deployment.yaml
```

### 9. Multi-Environment Deployment

```yaml
# multi-env.yaml
workflow:
  - build
  - deploy

build:
  image: docker:latest
  copyFiles: true
  script:
    - docker build -t myapp:latest .

deploy:
  image: alpine:latest
  condition: $DEPLOY_ENV
  script:
    - echo "Deploying to environment: $DEPLOY_ENV"
    - |
      if [ "$DEPLOY_ENV" = "staging" ]; then
        echo "Configuring staging environment..."
      elif [ "$DEPLOY_ENV" = "production" ]; then
        echo "Configuring production environment..."
      fi
```

## 📊 Monitoring and Debugging

### 10. Performance Monitoring

```yaml
# monitoring.yaml
workflow:
  - health-check
  - performance-test
  - log-analysis

health-check:
  image: alpine:latest
  script:
    - wget --spider http://localhost:8080/health || echo "Health check failed"

performance-test:
  image: alpine:latest
  script:
    - apk add --no-cache apache2-utils
    - ab -n 1000 -c 10 http://localhost:8080/api/test

log-analysis:
  image: alpine:latest
  script:
    - echo "Analyzing application logs..."
    - grep -i error /var/log/app.log || echo "No errors found"
```

### 11. Retry Mechanism Example

```yaml
# retry-example.yaml
workflow:
  - flaky-service
  - network-operation

flaky-service:
  image: alpine:latest
  retry:
    attempts: 3
    delay: 2
    backoff: 1.5
  script:
    - echo "Attempting to connect to flaky service..."
    - if [ $((RANDOM % 3)) -eq 0 ]; then exit 1; fi
    - echo "Connection successful!"

network-operation:
  image: alpine:latest
  retry:
    attempts: 5
    delay: 1
    backoff: 2.0
  script:
    - wget https://httpbin.org/delay/1 -O /tmp/response.json
    - cat /tmp/response.json
```

## 🔧 Advanced Usage

### 12. Custom Dockerfile Development Environment

```yaml
# dev-environment.yaml
workflow:
  - setup-env
  - run-dev

setup-env:
  dockerfile: "./dev.Dockerfile"
  copyFiles: true
  script:
    - echo "Development environment ready"

run-dev:
  image: setup-env-custom:latest
  copyFiles: true
  port:
    - "8080:8080"
    - "127.0.0.1:3000:3000"
  env:
    - NODE_ENV=development
    - DEBUG=true
  script:
    - npm run dev
```

### 13. Microservices Test Suite

```yaml
# microservices.yaml
workflow:
  - user-service
  - auth-service
  - api-gateway
  - integration-test

user-service:
  image: node:18-alpine
  copyFiles: true
  port:
    - "127.0.0.1:3001:3000"
  parallel: true
  script:
    - cd services/user-service
    - npm install
    - npm start

auth-service:
  image: node:18-alpine
  copyFiles: true
  port:
    - "127.0.0.1:3002:3000"
  parallel: true
  script:
    - cd services/auth-service
    - npm install
    - npm start

api-gateway:
  image: node:18-alpine
  copyFiles: true
  port:
    - "8080:8080"
  parallel: true
  script:
    - cd services/api-gateway
    - npm install
    - npm start

integration-test:
  image: node:18-alpine
  copyFiles: true
  script:
    - sleep 10  # Wait for services to start
    - cd tests
    - npm install
    - npm run integration-test
```

## 🎯 Tips and Best Practices

### Environment Variables
```bash
# Development
NODE_ENV=development pin apply -f app.yaml

# Production
NODE_ENV=production BRANCH=main pin apply -f app.yaml
```

### File Ignore Patterns
```yaml
copyIgnore:
  - "node_modules"
  - "*.log"
  - ".git"
  - "coverage"
  - "dist"
  - ".env*"
```

### Port Configuration Best Practices
```yaml
# Development - bind to localhost only
port:
  - "127.0.0.1:3000:3000"
  - "127.0.0.1:8080:8080"

# Production - bind to all interfaces
port:
  - "3000:3000"
  - "8080:8080"

# Specific IP binding
port:
  - "192.168.1.100:3000:3000"
```

## 🔄 Daemon Mode Examples

### 14. HTTP API Pipeline Triggering

Start daemon mode:
```bash
pin apply --daemon
```

Trigger pipelines via HTTP:
```bash
# Trigger a simple build pipeline
curl -X POST -H "Content-Type: application/yaml" \
  --data-binary @build.yaml \
  http://localhost:8081/trigger

# Monitor real-time events
curl -N http://localhost:8081/events
```

### 15. Production Monitoring Setup

```yaml
# production-monitor.yaml
workflow:
  - health-checks
  - metrics-collection
  - alerting

health-checks:
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -f http://app:8080/health
    - curl -f http://db:5432/ping

metrics-collection:
  image: prom/node-exporter:latest
  port:
    - "9100:9100"
  script:
    - echo "Metrics collection started"

alerting:
  image: alpine:latest
  condition: $ALERT_ENABLED == "true"
  script:
    - echo "Setting up alerting..."
    - curl -X POST http://alertmanager:9093/api/v1/alerts
```

These examples demonstrate how Pin can be used in various scenarios, from simple builds to complex production deployments with real-time monitoring.