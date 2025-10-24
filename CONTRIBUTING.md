# Contributing to Goqu-LINQ

Thank you for your interest in contributing to Goqu-LINQ! 🎉

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in [Issues](https://github.com/Natalieihs/goqu-linq/issues)
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Go version and OS information
   - Code samples if applicable

### Suggesting Enhancements

1. Check existing [Issues](https://github.com/Natalieihs/goqu-linq/issues) and [Pull Requests](https://github.com/Natalieihs/goqu-linq/pulls)
2. Create a new issue describing:
   - The enhancement and its benefits
   - Possible implementation approach
   - Examples of usage

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes with clear commit messages
4. Add tests for new functionality
5. Ensure all tests pass: `go test ./...`
6. Update documentation if needed
7. Push to your fork and submit a pull request

## Development Setup

```bash
# Clone your fork
git clone https://github.com/Natalieihs/goqu-linq.git
cd goqu-linq

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./core
```

## Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Run `go vet` to catch common mistakes
- Use meaningful variable and function names
- Add comments for exported functions and types

## Testing

- Write unit tests for new functionality
- Maintain or improve code coverage
- Test with multiple Go versions if possible

## Documentation

- Update README.md for user-facing changes
- Add/update code comments for exported APIs
- Include examples for new features

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
