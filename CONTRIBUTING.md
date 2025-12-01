# Contributing to Herald

Thank you for your interest in contributing to Herald! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

Before creating a bug report:
1. Check existing issues to avoid duplicates
2. Collect relevant information (version, OS, configuration)

When creating a bug report, include:
- Clear, descriptive title
- Steps to reproduce
- Expected behavior
- Actual behavior
- Configuration files (sanitized)
- Log output
- System information

### Suggesting Features

Feature requests are welcome! Please include:
- Clear description of the feature
- Use case and motivation
- Potential implementation approach
- Any alternatives considered

### Pull Requests

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/herald.git
   cd herald
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/my-feature
   # or
   git checkout -b fix/bug-description
   ```

3. **Make Changes**
   - Follow coding standards (see below)
   - Add tests for new functionality
   - Update documentation

4. **Test Your Changes**
   ```bash
   make fmt      # Format code
   make tidy     # Tidy dependencies
   make lint     # Run linters
   make security # Security checks
   make test     # Run tests
   make build    # Build binaries
   ```

5. **Commit**
   ```bash
   git add .
   git commit -m "feat: add new probe type"
   ```

   Use conventional commits:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation
   - `test:` - Tests
   - `refactor:` - Code refactoring
   - `perf:` - Performance improvements
   - `chore:` - Maintenance

6. **Push and Create PR**
   ```bash
   git push origin feature/my-feature
   ```
   Then create a pull request on GitHub.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- golangci-lint
- gosec
- goreleaser

### Install Dependencies

```bash
make setup
```

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./pkg/probe/...

# Run with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run security checks
make security

# Run all checks
make all
```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Write meaningful comments for exported functions
- Keep functions focused and small

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}

// Bad - loses error context
if err != nil {
    return err
}
```

### Logging

Use structured logging with zap:

```go
zap.S().Info("Starting probe", "type", "http", "target", target)
zap.S().Error("Probe failed", "error", err)
zap.S().Debug("Detailed info", "data", data)
```

### Testing

- Write tests for new functionality
- Aim for meaningful test coverage
- Use table-driven tests where appropriate

```go
func TestProbeHTTP(t *testing.T) {
    tests := []struct {
        name    string
        probe   *ProbeHTTP
        wantErr bool
    }{
        {
            name: "successful probe",
            probe: &ProbeHTTP{
                Host: "localhost",
                Port: 8080,
            },
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Project Structure

```
herald/
├── main.go                 # Entry point
├── pkg/
│   ├── bfd/               # BFD implementation
│   ├── config/            # Configuration
│   ├── probe/             # Health probes
│   ├── scheduler/         # Probe scheduling
│   ├── service/           # Service management
│   └── speaker/           # BGP speaker
├── docs/                  # Documentation
├── Makefile              # Build automation
└── README.md
```

## Adding New Features

### Adding a New Probe Type

1. Create `pkg/probe/probe_NEWTYPE.go`
2. Implement `ProbeInterface`
3. Add configuration struct
4. Update `probe.go` to support new type
5. Add tests
6. Update documentation

Example:

```go
package probe

type ProbeNEWTYPE struct {
    // Configuration fields
}

func (p *ProbeNEWTYPE) Run(ctx context.Context) (*ProbeStatus, error) {
    // Implementation
}
```

### Adding Configuration Options

1. Update struct in `pkg/config/config.go`
2. Add YAML tags
3. Document in `docs/configuration.md`
4. Add example in `docs/examples/`

## Documentation

- Update README.md for user-facing changes
- Update docs/ for detailed documentation
- Add examples for new features
- Keep documentation in sync with code

## Release Process

Releases are automated via GitHub Actions and goreleaser:

1. Update CHANGELOG.md
2. Create and push tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions builds and creates release

## Questions?

- Open an issue for questions
- Join discussions on GitHub Discussions
- Check existing documentation

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
