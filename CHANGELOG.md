# Changelog

All notable changes will start from this initial public release.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) and [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-11-29

### Added
- Initial public release of Raqeem platform (devices backend, mentor backend, frontends, PostgreSQL + MinIO integration)
- Remote command execution with full multi-line result display in Mentor dashboard
- Standardized container image naming: `ghcr.io/<owner>/raqeem/<component>`
- Comprehensive OpenAPI specs for both backends
- Interactive release script with optional git tag push (`scripts/tag-release.sh`)

### Quality
- 90%+ test coverage target established
- Linting (ruff, golangci-lint, ESLint) and type checking (mypy) configured

### Notes
- This is treated as the first release; prior pre-release references removed.

[1.0.0]: https://github.com/mj-nehme/raqeem/releases/tag/v1.0.0
