# Contributing to Mate

Thank you for your interest in contributing to Mate!

## Reporting Bugs

Open a [GitHub Issue](https://github.com/muralx/mate/issues) with:
- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS

## Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-change`)
3. Make your changes
4. Ensure all tests pass (`go test ./...`)
5. Submit a Pull Request

## Development Setup

```bash
git clone https://github.com/muralx/mate.git
cd mate

# Build
go build ./...

# Run tests
go test ./...

# Vet
go vet ./...

# Format
go fmt ./...
```

## Testing

All pull requests must include tests. Run the full suite before submitting:

```bash
go test -race ./...
```

## Code Style

- Run `go fmt` and `go vet` before committing
- Follow existing patterns in the codebase
- Keep files focused — one clear responsibility per file
- Write tests for all new functionality

## Commit Messages

Use conventional commit style:
- `feat: add new widget`
- `fix: correct focus cycling bug`
- `docs: update API reference`
- `chore: update dependencies`
