# Contributing to PromptGuard

Thank you for your interest in contributing to PromptGuard! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or higher
- Git
- API keys for testing (OpenAI, Anthropic, or Mistral)

### Development Setup
1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/promptguard.git
   cd promptguard
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o pg main.go
   ```
5. Run tests:
   ```bash
   go test ./...
   ```

## 🎯 How to Contribute

### Reporting Issues
- Use GitHub Issues to report bugs or request features
- Search existing issues before creating new ones
- Provide clear reproduction steps for bugs
- Include system information (OS, Go version)

### Submitting Changes
1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes
3. Write/update tests
4. Ensure all tests pass:
   ```bash
   go test ./...
   ```
5. Commit with a clear message:
   ```bash
   git commit -m "feat: add new assertion type for semantic similarity"
   ```
6. Push and create a Pull Request

### Code Guidelines
- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for complex logic
- Write unit tests for new functionality
- Update documentation as needed

### Commit Message Format
We use conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation updates
- `test:` - Test improvements
- `refactor:` - Code refactoring
- `perf:` - Performance improvements

## 🛠️ Development Areas

### High Priority
- **Additional LLM Providers**: Ollama, Azure OpenAI, Cohere
- **Enhanced Assertions**: Semantic similarity, fact-checking
- **UI Improvements**: Better diff viewer, interactive charts
- **Performance**: Caching, parallel execution optimizations

### Medium Priority
- **IDE Integrations**: VS Code extension, JetBrains plugin
- **Advanced Analytics**: Trend analysis, cost optimization
- **Security**: Enhanced prompt injection detection
- **Documentation**: Video tutorials, examples

### Architecture Overview
```
promptguard/
├── cmd/           # CLI commands (test, ci, view)
├── internal/      # Core packages
│   ├── config/    # Configuration management
│   ├── runner/    # Test execution engine
│   ├── providers/ # LLM provider clients
│   ├── assertions/# Assertion evaluators
│   ├── reporter/  # Output formatters
│   ├── metrics/   # Data storage
│   ├── github/    # CI/CD integrations
│   └── viewer/    # Web interface
├── prompts/       # Example prompts
└── artifacts/     # Generated reports
```

## 🧪 Testing

### Running Tests
```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests (requires API keys)
export OPENAI_API_KEY="sk-..."
go test -tags=integration ./...
```

### Test Categories
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test with real LLM providers
- **End-to-End Tests**: Test complete workflows

### Adding Tests
- Place tests in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Mock external dependencies when possible
- Test both success and error cases

## 📚 Documentation

### Updating Documentation
- Update README.md for user-facing changes
- Add docstrings for new public APIs
- Update examples in `/examples` directory
- Consider adding to wiki for detailed guides

### Documentation Standards
- Use clear, concise language
- Provide working code examples
- Include common use cases
- Add troubleshooting sections

## 🔄 Release Process

### Versioning
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Release Checklist
- [ ] Update version in relevant files
- [ ] Update CHANGELOG.md
- [ ] Tag release: `git tag v1.2.3`
- [ ] Build binaries for all platforms
- [ ] Update GitHub release with binaries
- [ ] Update documentation

## 🤝 Community

### Communication
- **GitHub Discussions**: General questions and ideas
- **Issues**: Bug reports and feature requests
- **Discord**: Real-time chat and support
- **Twitter**: Updates and announcements

### Code of Conduct
- Be respectful and inclusive
- Help others learn and grow
- Provide constructive feedback
- Follow the [Contributor Covenant](https://www.contributor-covenant.org/)

## 📋 Pull Request Template

When creating a PR, please include:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Other (please describe)

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project conventions
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or marked as such)
```

## 🎉 Recognition

Contributors will be:
- Listed in the project README
- Credited in release notes
- Invited to join the core team (for significant contributions)
- Eligible for PromptGuard swag and recognition

Thank you for helping make PromptGuard better! 🚀
