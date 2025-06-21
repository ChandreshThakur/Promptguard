# Changelog

All notable changes to PromptGaurd by Chandresh will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial PromptGaurd by Chandresh implementation
- CLI commands: `pg test`, `pg ci`, `pg view`
- OpenAI provider support
- Core assertion types: answer-relevance, contains-json, cost, toxicity
- Multiple output formats: console, JSON, JUnit XML, HTML, Markdown
- Interactive web viewer for test results
- GitHub Actions integration
- SQLite metrics storage
- Baseline comparison for regression detection

### Coming Soon
- Anthropic and Mistral provider support
- LLM-rubric and closed-QA assertions
- Enhanced diff viewer with syntax highlighting
- Plugin SDK for custom assertions

## [0.1.0] - 2024-12-21

### Added
- 🎉 Initial release of PromptGuard
- ⚡ Core testing framework for LLM prompts
- 🔧 CLI with `test`, `ci`, and `view` commands
- 🤖 OpenAI provider integration (GPT-4, GPT-3.5-turbo)
- 📊 Rich assertion system (relevance, JSON validation, cost limits)
- 📋 Multiple report formats (HTML, JSON, JUnit, Markdown)
- 🌐 Interactive web viewer for exploring results
- 🚀 GitHub Actions integration with annotations
- 💾 SQLite-based metrics storage for historical tracking
- 📈 Baseline comparison for regression detection
- 🎯 Template-based prompt system with variable substitution
- 🛡️ Basic toxicity detection
- 💰 Cost tracking and budget enforcement

### Technical Details
- Go 1.21+ single-binary distribution
- Cross-platform support (Linux, macOS, Windows)
- Parallel test execution
- Configurable via YAML
- Environment variable support for API keys
- Comprehensive error handling and logging

### Example Configuration
```yaml
description: "E-commerce prompt tests"
prompts:
  - prompts/*.prompt
providers:
  - id: openai:gpt-4o
    config:
      temperature: 0
tests:
  - vars: {customer: "Alice", product: "Pro"}
    assert:
      - type: answer-relevance
        value: "Pro Plan benefits"
        threshold: 0.7
      - type: cost
        threshold: 0.003
```

### Documentation
- Comprehensive README with quick start guide
- Example prompts and configuration
- GitHub Actions workflow templates
- Contributing guidelines
- MIT license

---

**Full Changelog**: https://github.com/promptguard/promptguard/commits/v0.1.0
