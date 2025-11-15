# Contributing to Raqeem

Thank you for your interest in contributing to Raqeem! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and professional in all interactions.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/mj-nehme/raqeem/issues)
2. Create a new issue with:
   - Clear, descriptive title
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, versions)
   - Relevant logs or screenshots

### Suggesting Enhancements

1. Check existing issues for similar suggestions
2. Create a new issue with:
   - Clear description of the enhancement
   - Use cases and benefits
   - Potential implementation approach

### Pull Requests

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/raqeem.git
   cd raqeem
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make Changes**
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation as needed
   - Keep commits atomic and well-described

4. **Test Locally**
   ```bash
   # Run all tests
   pytest tests/
   cd devices/backend/src && pytest
   cd mentor/backend/src && go test ./...
   cd devices/frontend && npm test
   cd mentor/frontend && npm test
   
   # Check linting
   cd devices/backend/src && ruff check .
   cd mentor/backend/src && golangci-lint run
   cd devices/frontend && npm run lint
   cd mentor/frontend && npm run lint
   ```

5. **Commit Changes**
   ```bash
   git add .
   git commit -m "type: brief description
   
   Longer explanation of what changed and why."
   ```
   
   Commit types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

6. **Push and Create PR**
   ```bash
   git push origin your-branch-name
   ```
   Then create a PR on GitHub with:
   - Clear title and description
   - Link to related issues
   - Summary of changes
   - Testing performed

## Development Setup

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed setup instructions.

### Quick Start

```bash
# Install dependencies
./start.sh

# Run tests
pytest
go test ./...
npm test
```

## Code Style

### Python
- Follow PEP 8
- Use type hints
- Maximum line length: 120
- Format with ruff
- Use async/await for I/O operations

### Go
- Follow Go conventions
- Use gofmt
- Document exported functions
- Handle errors explicitly

### JavaScript/React
- Use modern ES6+ syntax
- Follow ESLint configuration
- Use functional components with hooks
- Document complex logic

## Testing

- Write tests for all new features
- Maintain or improve code coverage
- Include unit, integration, and E2E tests as appropriate
- Mock external dependencies

## Documentation

- Update relevant docs in `/docs` directory
- Keep README.md current
- Document API changes
- Add inline comments for complex logic
- Update CHANGELOG.md for user-facing changes

## Review Process

1. All PRs require review before merging
2. CI must pass (tests, linting, build)
3. Maintain backwards compatibility when possible
4. Address review feedback promptly

## Release Process

See [VERSION_MANAGEMENT.md](docs/VERSION_MANAGEMENT.md) for release procedures.

## Questions?

- Check [Documentation](docs/)
- Search [Issues](https://github.com/mj-nehme/raqeem/issues)
- Create a new issue with the `question` label

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
