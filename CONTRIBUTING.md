# Contributing to go-anki-deck

Thank you for your interest in contributing to go-anki-deck! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions.

## How to Contribute

### Reporting Issues

- Check if the issue already exists
- Include Go version, OS, and relevant details
- Provide minimal reproducible examples
- Use clear, descriptive titles

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add or update tests as needed
5. Ensure all tests pass (`make test`)
6. Run the linter (`make lint`)
7. Commit your changes with clear messages
8. Push to your fork
9. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/go-anki-deck.git
cd go-anki-deck

# Install dependencies
go mod download

# Run tests
make test

# Run benchmarks
make bench

# Generate coverage report
make coverage
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write clear, self-documenting code
- Add comments for exported functions
- Keep functions focused and small

### Testing

- Write tests for new features
- Maintain or improve code coverage
- Include both positive and negative test cases
- Use table-driven tests where appropriate

### Documentation

- Update README.md for new features
- Add godoc comments for exported types and functions
- Include examples in documentation

## Questions?

Feel free to open an issue for any questions about contributing.