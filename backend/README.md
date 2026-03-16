# Go Fiber Starter

A Go API project starter template based on the [Fiber](https://github.com/gofiber/fiber) framework, designed for rapid development and high-performance API services.

_[中文文档](README_zh.md)_

## Features

- 🚀 Built on Go Fiber framework, offering extremely fast HTTP performance
- 📝 Integrated Swagger documentation for clear API visibility
- 🔐 Built-in JWT authentication system
- 📦 SQLite database for simplicity and ease of use
- 🔄 Automatic database migration functionality
- 📊 Elegant logging mechanism
- 🛠️ Complete error handling middleware
- 🐳 Docker support for one-click deployment

## 项目结构

```
go-fiber-starter/
├── cmd/                     # Application entry points
│   ├── api.go               # API service configuration
│   └── main.go              # Main program entry
├── config/                  # Configuration files
│   └── config.yaml          # Application configuration
├── data/                    # Data storage
│   └── db.sqlite            # SQLite database file
├── docs/                    # Swagger documentation
│   ├── docs.go              # Auto-generated documentation code
│   ├── swagger.json         # Swagger JSON configuration
│   └── swagger.yaml         # Swagger YAML configuration
├── internal/                # Internal application code
│   ├── api/                 # API handlers
│   │   ├── auth/            # Authentication-related API
│   │   │   ├── handler.go   # Authentication handler functions
│   │   │   └── router.go    # Authentication routes
│   │   └── response/        # Response handling
│   │       └── response.go  # Response utility functions
│   ├── middleware/          # Middleware
│   │   └── middleware.go    # Global middleware
│   ├── model/               # Data models
│   │   ├── base/            # Base models
│   │   │   └── base.go      # Model base class
│   │   └── user/            # User model
│   │       └── user.go      # User struct
│   └── service/             # Business logic layer
│       └── user.go          # User service
├── log/                     # Log files
│   └── log.json             # JSON format logs
├── scripts/                 # Helper scripts (Windows)
│   ├── build.bat            # Build binary
│   ├── run.bat              # Run API server
│   └── test.bat             # Run tests
├── pkg/                     # Public packages
│   ├── config/              # Configuration processing
│   │   └── config.go        # Configuration loading logic
│   ├── db/                  # Database operations
│   │   ├── db.go            # Database connection
│   │   ├── migrate.go       # Database migration
│   │   └── user.go          # User database operations
│   ├── logger/              # Log processing
│   │   └── logger.go        # Log configuration
│   └── util/                # Utility functions
│       └── file.go          # File operation utilities
├── .dockerignore            # Docker ignore file
├── docker-compose.yml       # Docker Compose configuration
├── Dockerfile               # Docker build file
├── go.mod                   # Go module file
├── go.sum                   # Go dependency verification
└── README.md                # Project documentation
```

## Quick Start

### Prerequisites

1. Install [Go](https://golang.org/dl/) (version 1.24 or higher)
2. Clone this repository

```bash
git clone https://github.com/your-username/go-fiber-starter.git
cd go-fiber-starter
```

### Local Running

1. Install dependencies

```bash
go mod download
```

2. Run the application

```bash
go run ./cmd
```

3. Access the application

The API service runs by default at `http://localhost:25610`

Swagger documentation can be accessed via `http://localhost:25610/swagger/`

### Windows Scripts

If you are on Windows, you can use the scripts under `scripts/`:

```bat
scripts\build.bat
scripts\run.bat
scripts\test.bat
```

### Running Tests

```bash
go test ./...
```

The auth HTTP tests use an in-memory SQLite database and do not touch `data/db.sqlite`.

### Running with Docker

1. Build and start the container

```bash
docker-compose up -d
```

2. Access the application

The API service runs by default at `http://localhost:25610`

## API Documentation

This project uses Swagger to automatically generate API documentation. After starting the application, visit the `/swagger/` path to view the complete API documentation.

## Main API Endpoints

- **Authentication Related**

  - `POST /register` - User registration
  - `POST /login` - User login

- **User Related**
  - `GET /api/user/profile` - Get user profile (requires authentication)

## Configuration

The configuration file is located at `config/config.yaml`, with main configuration items including:

```yaml
app:
  port: "25610" # Application port
  env: "development" # Environment setting (development/production)
jwt:
  secret: "your-secret" # JWT key (environment variables recommended for production)
  expiration: 86400 # Token validity period (seconds)
database:
  path: "data/db.sqlite" # SQLite database path
```

## Directory Structure Description

- `cmd/`: Application entry points
- `config/`: Configuration files
- `docs/`: Swagger documentation
- `internal/`: Internal application code, not exposed externally
  - `api/`: API handlers and routes
  - `middleware/`: Middleware
  - `model/`: Data models
  - `service/`: Business logic
- `pkg/`: Public packages, can be referenced externally
  - `config/`: Configuration processing
  - `db/`: Database operations
  - `logger/`: Log processing
  - `util/`: Utility functions

## Docker Deployment

The project provides Docker deployment-related files:

- `Dockerfile`: For building Docker images
- `docker-compose.yml`: For Docker Compose deployment
- `.dockerignore`: Excludes unnecessary files

For detailed Docker deployment instructions, please refer to [docker-readme.md](docker-readme.md).

## Development Guide

### Adding New Routes

1. Create a new package under `internal/api`
2. Implement handler functions
3. Register routes in `cmd/api.go`

### Adding New Models

1. Create a new package and model file under `internal/model`
2. Add the model to the automatic migration list in `pkg/db/migrate.go`

### Generating Swagger Documentation

Use the [swag](https://github.com/swaggo/swag) tool to update API documentation:

```bash
# Install swag tool
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/main.go
```

## Contribution Guidelines

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Submit a Pull Request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

Copyright © 2025 ydfk.

## Contact Information

For any questions or suggestions, please contact:

- Project maintainer: ydfk
- Email: [lyh6728326@gmail.com]
