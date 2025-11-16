# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-11-16

### Changed
- **Removed DockerHub integration**: Removed automatic Docker image push to DockerHub from CI pipeline
- **Improved reliability**: Enhanced test coverage and end-to-end testing for production readiness

### Added
- Comprehensive Swagger/OpenAPI documentation for both backends:
  - Devices Backend: FastAPI auto-generated docs at `/docs` and `/redoc`
  - Mentor Backend: Swagger UI at `/swagger/index.html` and `/docs` redirect

### Technical Details
- **CI/CD**: Removed DockerHub push job, keeping local Docker builds for validation
- **Test Coverage Target**: Maintaining 90% coverage threshold across all components
- **Documentation**: Both backends now have consistent API documentation via Swagger/OpenAPI

### Infrastructure
- CI pipeline now only builds Docker images for validation without pushing to registry
- Maintained test coverage reporting to Codecov for all four components

## [0.1.0] - 2025-11-15

### Added
- Initial release of Raqeem monitoring system
- Device monitoring backend (Python/FastAPI)
- Mentor dashboard backend (Go)
- Device simulator frontend (React)
- Mentor dashboard frontend (React)
- Real-time device metrics collection
- Device activity tracking
- Screenshot capture and storage
- Alert system for critical events
- Remote command execution
- PostgreSQL database persistence
- S3-compatible storage for screenshots
- Comprehensive test coverage (backend and frontend)
- CI/CD pipeline with GitHub Actions
- Docker deployment support

### Technical Details
- **Test Coverage**: 189 backend unit tests, 76 devices frontend tests, 23 mentor frontend tests, 15 integration tests
- **Linting**: Clean ruff (Python), golangci-lint (Go), ESLint (JavaScript)
- **Type Checking**: mypy configured with relaxed settings for rapid development
- **CI/CD**: GitHub Actions with Docker build, test, and deployment automation
- **Documentation**: Comprehensive guides (9,960+ lines across 16 docs)

### Security
- UUID-based device identification
- Strict validation for all API inputs
- CORS configuration for frontend security
- Token-based authentication support
- No exposed secrets or credentials

[0.2.0]: https://github.com/mj-nehme/raqeem/releases/tag/v0.2.0
[0.1.0]: https://github.com/mj-nehme/raqeem/releases/tag/v0.1.0
