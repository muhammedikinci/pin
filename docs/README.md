# 📚 Pin Documentation

Welcome to the comprehensive documentation for Pin - a local pipeline tool with Docker integration and daemon mode support.

## 📖 Documentation Index

### 🚀 Getting Started
- [Main README](../README.md) - Overview, installation, and quick start
- [Examples](examples.md) - Practical examples for different scenarios
- [Use Cases](use-cases.md) - Real-world applications and workflows

### 📋 Reference
- [API Reference](api-reference.md) - Complete HTTP API documentation for daemon mode
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

## 🎯 Quick Navigation

### For Beginners
1. Start with the [Main README](../README.md) to understand what Pin is
2. Check out [Examples](examples.md) for basic usage patterns
3. Reference [Troubleshooting](troubleshooting.md) if you encounter issues

### For Advanced Users
1. Explore [Use Cases](use-cases.md) for complex workflow ideas
2. Use [API Reference](api-reference.md) for daemon mode integration
3. Check [Examples](examples.md) for advanced configuration patterns

### For Integrators
1. Study [API Reference](api-reference.md) for HTTP API integration
2. Review [Use Cases](use-cases.md) for CI/CD integration patterns
3. Reference [Examples](examples.md) for automation scripts

## 🔧 Core Concepts

### Pipeline Configuration
Pin uses YAML configuration files to define workflows:

```yaml
workflow:
  - build
  - test
  - deploy

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

deploy:
  image: alpine:latest
  condition: $BRANCH == "main"
  script:
    - echo "Deploying to production"
```

### Key Features
- **Local Execution**: Run pipelines on your local machine
- **Docker Integration**: Consistent environments using containers
- **Daemon Mode**: Long-running service with HTTP API
- **Real-time Monitoring**: Server-Sent Events for live pipeline updates
- **Conditional Execution**: Run jobs based on environment conditions
- **Retry Mechanism**: Automatic retry with exponential backoff
- **Parallel Jobs**: Execute multiple jobs simultaneously

## 🌟 Popular Examples

### Quick Start
```bash
# Install Pin
go install github.com/muhammedikinci/pin/cmd/cli@latest

# Run a simple pipeline
pin apply -f pipeline.yaml

# Start daemon mode
pin apply --daemon
```

### Common Use Cases
- **Development**: Consistent local development environments
- **Testing**: Automated testing across different configurations
- **CI/CD**: Build, test, and deployment automation
- **Data Processing**: ETL pipelines and batch jobs

## 🤝 Contributing

Pin is an open-source project. Contributions are welcome!

- [GitHub Repository](https://github.com/muhammedikinci/pin)
- [Issues](https://github.com/muhammedikinci/pin/issues)
- [Discussions](https://github.com/muhammedikinci/pin/discussions)

## 📞 Support

Need help? Check these resources:

1. **Documentation**: You're here! Browse the guides above
2. **Examples**: Practical examples in the [examples](examples.md) section
3. **Troubleshooting**: Common issues in [troubleshooting](troubleshooting.md)
4. **GitHub Issues**: Report bugs or ask questions
5. **Discussions**: Community discussions and feature requests

---

**Happy pipelining with Pin! 🔥**