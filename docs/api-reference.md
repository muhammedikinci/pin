# 🌐 Pin API Reference

This document provides comprehensive documentation for Pin's HTTP API when running in daemon mode.

## 🚀 Getting Started

Start Pin in daemon mode to enable the HTTP API:

```bash
# Start daemon without initial pipeline
pin apply --daemon

# Start daemon with specific pipeline
pin apply --daemon -f pipeline.yaml

# Start daemon on custom port
pin apply --daemon --port 8082
```

**Base URL**: `http://localhost:8081` (default)

## 📋 API Endpoints

### 1. Root Endpoint

Get API information and available endpoints.

**Endpoint**: `GET /`

**Response**:
```json
{
  "service": "pin-daemon",
  "version": "1.0.0",
  "status": "running",
  "endpoints": {
    "/": "API information",
    "/health": "Health check and status",
    "/events": "Server-Sent Events stream",
    "/trigger": "Trigger pipeline execution"
  },
  "uptime": "2h 15m 30s",
  "connected_clients": 3
}
```

**Example**:
```bash
curl http://localhost:8081/
```

### 2. Health Check

Check daemon health and connected client count.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "healthy",
  "uptime": "2h 15m 30s",
  "connected_clients": 3,
  "last_pipeline": "2024-01-15T10:30:00Z",
  "total_pipelines_executed": 42
}
```

**Example**:
```bash
curl http://localhost:8081/health
```

### 3. Server-Sent Events Stream

Connect to real-time event stream for pipeline monitoring.

**Endpoint**: `GET /events`

**Response**: Server-Sent Events stream

**Event Types**:
- `daemon_start`: Service started
- `pipeline_trigger`: Pipeline execution requested
- `job_container_start`: Container started for job
- `log`: Real-time log messages
- `job_completed`: Job finished successfully
- `job_failed`: Job failed with error
- `pipeline_complete`: Pipeline execution finished
- `daemon_stop`: Service shutting down

**Example**:
```bash
# Command line
curl -N http://localhost:8081/events

# JavaScript
const eventSource = new EventSource('http://localhost:8081/events');
eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log(`[${data.level}] ${data.message}`);
};
```

**Sample Events**:
```
data: {"level":"info","message":"Pipeline execution started","job":"build","timestamp":"2024-01-15T10:30:00Z"}

data: {"level":"info","message":"Container started","job":"build","container_id":"abc123","timestamp":"2024-01-15T10:30:01Z"}

data: {"level":"success","message":"Job completed successfully","job":"build","duration":"45s","timestamp":"2024-01-15T10:30:45Z"}

data: {"level":"info","message":"Pipeline execution completed","total_duration":"1m 30s","timestamp":"2024-01-15T10:31:30Z"}
```

### 4. Trigger Pipeline

Execute a pipeline by sending YAML configuration.

**Endpoint**: `POST /trigger`

**Content-Type**: `application/yaml`

**Request Body**: Pipeline YAML configuration

**Response**:
```json
{
  "status": "triggered",
  "message": "Pipeline execution started",
  "pipeline_id": "pipeline-abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Example**:
```bash
# Trigger with file
curl -X POST \
  -H "Content-Type: application/yaml" \
  --data-binary @pipeline.yaml \
  http://localhost:8081/trigger

# Trigger with inline YAML
curl -X POST \
  -H "Content-Type: application/yaml" \
  -d 'workflow:
  - hello
hello:
  image: alpine:latest
  script:
    - echo "Hello from API!"' \
  http://localhost:8081/trigger
```

## 📡 Real-time Monitoring Examples

### JavaScript Client

```javascript
class PinMonitor {
  constructor(baseUrl = 'http://localhost:8081') {
    this.baseUrl = baseUrl;
    this.eventSource = null;
  }

  // Connect to event stream
  connect() {
    this.eventSource = new EventSource(`${this.baseUrl}/events`);

    this.eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleEvent(data);
    };

    this.eventSource.onerror = (error) => {
      console.error('EventSource failed:', error);
    };
  }

  // Handle incoming events
  handleEvent(data) {
    const { level, message, job, timestamp } = data;

    switch (level) {
      case 'info':
        console.log(`ℹ️ [${job || 'system'}] ${message}`);
        break;
      case 'success':
        console.log(`✅ [${job}] ${message}`);
        break;
      case 'error':
        console.error(`❌ [${job}] ${message}`);
        break;
      default:
        console.log(`📝 [${job || 'system'}] ${message}`);
    }
  }

  // Trigger pipeline
  async triggerPipeline(yamlConfig) {
    try {
      const response = await fetch(`${this.baseUrl}/trigger`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/yaml'
        },
        body: yamlConfig
      });

      const result = await response.json();
      console.log('Pipeline triggered:', result);
      return result;
    } catch (error) {
      console.error('Failed to trigger pipeline:', error);
    }
  }

  // Get health status
  async getHealth() {
    try {
      const response = await fetch(`${this.baseUrl}/health`);
      return await response.json();
    } catch (error) {
      console.error('Failed to get health:', error);
    }
  }

  // Disconnect from event stream
  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
  }
}

// Usage
const monitor = new PinMonitor();
monitor.connect();

// Trigger a pipeline
const pipeline = `
workflow:
  - test
test:
  image: alpine:latest
  script:
    - echo "Hello from JavaScript!"
`;

monitor.triggerPipeline(pipeline);
```

### Python Client

```python
import requests
import sseclient
import json
import yaml

class PinClient:
    def __init__(self, base_url='http://localhost:8081'):
        self.base_url = base_url

    def get_health(self):
        """Get daemon health status"""
        response = requests.get(f'{self.base_url}/health')
        return response.json()

    def trigger_pipeline(self, pipeline_config):
        """Trigger pipeline execution"""
        if isinstance(pipeline_config, dict):
            yaml_content = yaml.dump(pipeline_config)
        else:
            yaml_content = pipeline_config

        response = requests.post(
            f'{self.base_url}/trigger',
            headers={'Content-Type': 'application/yaml'},
            data=yaml_content
        )
        return response.json()

    def stream_events(self, callback=None):
        """Stream real-time events"""
        response = requests.get(f'{self.base_url}/events', stream=True)
        client = sseclient.SSEClient(response)

        for event in client.events():
            data = json.loads(event.data)

            if callback:
                callback(data)
            else:
                self.default_event_handler(data)

    def default_event_handler(self, data):
        """Default event handler"""
        level = data.get('level', 'info')
        message = data.get('message', '')
        job = data.get('job', 'system')

        emoji = {
            'info': 'ℹ️',
            'success': '✅',
            'error': '❌',
            'warning': '⚠️'
        }.get(level, '📝')

        print(f"{emoji} [{job}] {message}")

# Usage example
client = PinClient()

# Check health
health = client.get_health()
print(f"Daemon status: {health['status']}")

# Trigger pipeline
pipeline = {
    'workflow': ['hello'],
    'hello': {
        'image': 'alpine:latest',
        'script': ['echo "Hello from Python!"']
    }
}

result = client.trigger_pipeline(pipeline)
print(f"Pipeline triggered: {result['pipeline_id']}")

# Stream events
client.stream_events()
```

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "bufio"
)

type PinClient struct {
    BaseURL string
    Client  *http.Client
}

type HealthResponse struct {
    Status           string `json:"status"`
    Uptime          string `json:"uptime"`
    ConnectedClients int    `json:"connected_clients"`
}

type TriggerResponse struct {
    Status     string `json:"status"`
    Message    string `json:"message"`
    PipelineID string `json:"pipeline_id"`
}

type Event struct {
    Level     string `json:"level"`
    Message   string `json:"message"`
    Job       string `json:"job,omitempty"`
    Timestamp string `json:"timestamp"`
}

func NewPinClient(baseURL string) *PinClient {
    return &PinClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *PinClient) GetHealth() (*HealthResponse, error) {
    resp, err := c.Client.Get(c.BaseURL + "/health")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var health HealthResponse
    err = json.NewDecoder(resp.Body).Decode(&health)
    return &health, err
}

func (c *PinClient) TriggerPipeline(yamlConfig string) (*TriggerResponse, error) {
    req, err := http.NewRequest("POST", c.BaseURL+"/trigger",
        bytes.NewBufferString(yamlConfig))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/yaml")

    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var trigger TriggerResponse
    err = json.NewDecoder(resp.Body).Decode(&trigger)
    return &trigger, err
}

func (c *PinClient) StreamEvents(eventHandler func(Event)) error {
    resp, err := c.Client.Get(c.BaseURL + "/events")
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()

        if len(line) > 6 && line[:6] == "data: " {
            var event Event
            if err := json.Unmarshal([]byte(line[6:]), &event); err == nil {
                eventHandler(event)
            }
        }
    }

    return scanner.Err()
}

func main() {
    client := NewPinClient("http://localhost:8081")

    // Check health
    health, err := client.GetHealth()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Daemon status: %s\n", health.Status)

    // Trigger pipeline
    pipeline := `
workflow:
  - hello
hello:
  image: alpine:latest
  script:
    - echo "Hello from Go!"
`

    result, err := client.TriggerPipeline(pipeline)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Pipeline triggered: %s\n", result.PipelineID)

    // Stream events
    client.StreamEvents(func(event Event) {
        emoji := map[string]string{
            "info":    "ℹ️",
            "success": "✅",
            "error":   "❌",
            "warning": "⚠️",
        }

        e, ok := emoji[event.Level]
        if !ok {
            e = "📝"
        }

        job := event.Job
        if job == "" {
            job = "system"
        }

        fmt.Printf("%s [%s] %s\n", e, job, event.Message)
    })
}
```

## 🔒 Authentication and Security

Currently, Pin daemon runs without authentication. For production use, consider:

### Reverse Proxy with Authentication

```nginx
# nginx.conf
server {
    listen 80;

    location / {
        auth_basic "Pin API";
        auth_basic_user_file /etc/nginx/.htpasswd;

        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /events {
        auth_basic "Pin API";
        auth_basic_user_file /etc/nginx/.htpasswd;

        proxy_pass http://localhost:8081/events;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # SSE specific headers
        proxy_set_header Cache-Control no-cache;
        proxy_buffering off;
        proxy_read_timeout 24h;
    }
}
```

### Network Security

```bash
# Bind to localhost only
pin apply --daemon --host 127.0.0.1

# Use firewall to restrict access
sudo ufw allow from 192.168.1.0/24 to any port 8081
```

## 📊 Error Handling

### HTTP Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid YAML or request format
- `500 Internal Server Error`: Pipeline execution error
- `503 Service Unavailable`: Daemon not ready

### Error Response Format

```json
{
  "error": "validation_failed",
  "message": "Pipeline validation failed: either 'image' or 'dockerfile' must be specified",
  "details": {
    "job": "build",
    "field": "image"
  }
}
```

## 🚀 Production Deployment

### Systemd Service

```ini
# /etc/systemd/system/pin-daemon.service
[Unit]
Description=Pin Pipeline Daemon
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=pin
WorkingDirectory=/opt/pin
ExecStart=/usr/local/bin/pin apply --daemon --host 127.0.0.1
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o pin ./cmd/cli/.

FROM alpine:latest
RUN apk --no-cache add docker-cli
WORKDIR /app
COPY --from=builder /app/pin .
EXPOSE 8081
CMD ["./pin", "apply", "--daemon", "--host", "0.0.0.0"]
```

This API reference provides comprehensive documentation for integrating with Pin's daemon mode, enabling programmatic pipeline execution and real-time monitoring.