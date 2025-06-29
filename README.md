# go-envsync

![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A minimal yet extensible Go library and CLI tool for unified environment variable and secrets management across multiple sources.

## Features

### ✅ Phase 1 - Core Foundation (Completed)
- **Multiple Provider Support**: Load configuration from various sources
- **JSON Schema Validation**: Strong type checking and validation
- **Multi-Format Export**: Export to JSON, YAML, and .env formats
- **CLI Interface**: User-friendly command-line tool
- **SDK Library**: Programmatic access for Go applications

### ✅ Phase 2 - Provider Ecosystem (Completed)
- **Provider Registry**: Dynamic provider registration system
- **Local File Provider**: Full support for .env files
- **Kubernetes Provider**: Stub implementation (ready for k8s dependencies)
- **Vault Provider**: Stub implementation (ready for HashiCorp Vault)
- **Provider Management**: List, filter, and configure providers

### 🚧 Future Phases
- **Phase 3**: Enhanced Validation & Security
- **Phase 4**: Advanced CLI Features
- **Phase 5**: Enterprise Features
- **Phase 6**: Ecosystem Integration

## Quick Start

### Installation

```bash
# Build from source
git clone https://github.com/Gosayram/go-envsync.git
cd go-envsync
make build

# Or use go install
go install github.com/Gosayram/go-envsync/cmd/go-envsync@latest
```

### Basic Usage

```bash
# List available providers
go-envsync providers

# Load from .env file
go-envsync load --from=.env

# Load with validation and export
go-envsync load --from=.env --validate=schema.json --export=json:config.json

# Load from multiple sources
go-envsync load --from=.env --from=local:.env.local --merge-strategy=override
```

### Example Configuration

Create a `.env` file:
```env
# Application settings
APP_NAME=my-application
APP_VERSION=1.0.0
APP_ENV=development

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp_db
```

Create a JSON schema (`.envschema.json`):
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "APP_NAME": {
      "type": "string",
      "pattern": "^[a-zA-Z0-9_-]+$"
    },
    "APP_VERSION": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$"
    }
  },
  "required": ["APP_NAME", "APP_VERSION"]
}
```

## Supported Providers

| Provider | Status | Description |
|----------|---------|-------------|
| **local** | ✅ Available | Load from local .env files |
| **kubernetes** | 🚧 Stub | Kubernetes Secrets/ConfigMaps (requires k8s deps) |
| **vault** | 🚧 Stub | HashiCorp Vault secrets (requires Vault deps) |
| **s3** | 📋 Planned | AWS S3 objects |

### Provider Usage

```bash
# List providers with details
go-envsync providers --details

# Filter providers
go-envsync providers --filter=local

# Load from different providers (when implemented)
go-envsync load --from=local:.env
go-envsync load --from=k8s:namespace/secret/my-secret
go-envsync load --from=vault:path/to/secret
```

## Library Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/Gosayram/go-envsync/pkg/client"
    "github.com/zahongirrahmankulov/go-envsync/pkg/client"
    "github.com/zahongirrahmankulov/go-envsync/pkg/providers/local"
)

func main() {
    client := client.New()
    
    // Add local provider
    client.AddProvider("local", local.NewProvider())
    
    // Load and validate
    env, err := client.Load(client.LoadOptions{
        Sources: []string{"local:.env"},
        Schema:  "./schema.json",
    })
    if err != nil {
        panic(err)
    }
    
    // Export to different formats
    env.ExportJSON("config.json")
}
```

## Supported Providers

| Provider   | Status    | Description                         |
| ---------- | --------- | ----------------------------------- |
| Local      | ✅ Planned | Local .env files with dotenv syntax |
| Kubernetes | ✅ Planned | K8s Secrets and ConfigMaps          |
| Vault      | ✅ Planned | HashiCorp Vault integration         |
| AWS S3     | ✅ Planned | S3 bucket configuration files       |

## Development

### Requirements

- Go 1.24.2+
- Make

### Building

```bash
# Install dependencies
make deps

# Build binary
make build

# Run tests
make test

# Run all quality checks
make check-all
```

### Project Structure

```
go-envsync/
├── cmd/envsync/          # CLI application
├── pkg/                  # Public API packages
│   ├── client/          # Main client interface
│   ├── providers/       # Provider implementations
│   ├── validator/       # Validation engine
│   └── exporter/        # Export functionality
├── internal/            # Internal packages
├── examples/            # Usage examples
├── docs/                # Documentation
└── scripts/             # Build scripts
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and run `make check-all`
5. Submit a pull request

## License

Apache License version 2.0 - see [LICENSE](LICENSE) for details.

## Roadmap

### Phase 1: Core Foundation
- [x] Project structure and build system
- [ ] Basic provider interface
- [ ] Local file provider implementation
- [ ] JSON Schema validation
- [ ] CLI basic commands

### Phase 2: Remote Providers
- [ ] Kubernetes Secrets integration
- [ ] HashiCorp Vault provider
- [ ] AWS S3 provider implementation
- [ ] Multi-source configuration loading

### Phase 3: Advanced Features
- [ ] Configuration merging and precedence
- [ ] Custom validation rules engine
- [ ] Encryption and secure export
- [ ] Performance optimizations

See [IDEA.md](IDEA.md) for detailed project vision and architecture. 