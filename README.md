# PromptGaurd by Chandresh

[![PromptGaurd](https://img.shields.io/badge/PromptGaurd-passing-brightgreen)](https://github.com/promptguard/promptguard)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

> **Continuous Integration Tests for LLM Prompts**  
> Give every team that stores prompts in Git a "unit-test runner for LLMs" so regressions in wording, temperature, or model version break the build before they reach production.

## 🎯 **MVP - Instant CI for Your Prompts**

**Pain Solved:** Teams ship prompt files (*.prompt / Markdown) yet have no automated regression tests; outputs drift after every model version upgrade.

### ⚡ **Quick Demo**

```bash
# 1. Add promptguard.yaml to your repo
echo "description: 'My prompt tests'
prompts: ['*.prompt']
providers: [{id: 'openai:gpt-4o', config: {temperature: 0}}]
tests: [{vars: {user: 'Alice'}, assert: [{type: 'cost', threshold: 0.01}]}]" > promptguard.yaml

# 2. Run tests locally  
pg test

# 3. Add to GitHub Actions
- uses: promptguard/run@v1
  with:
    openai-api-key: ${{ secrets.OPENAI_API_KEY }}
```

**Result:** Build fails on prompt drift with red/green markdown diffs! 🔴🟢

![Demo GIF Placeholder](https://via.placeholder.com/600x300/333/fff?text=Red/Green+Diff+Demo+GIF)

### 🏆 **Star Magnets**
- ✅ **Instantly pluggable** into existing CI (GitHub Actions ready)
- ✅ **YAML spec** = prompt + expected rubric
- ✅ **OpenAI + Ollama** support (local & cloud)
- ✅ **Markdown diff viewer** for failed assertions  
- ✅ **Red/green diffs** show exactly what changed
- ✅ **Taps AI toolchain trend** - prompts as first-class code

## 🚀 Quick Start

### Installation

**Option 1: Download Binary**
```bash
# Linux/macOS
curl -L https://github.com/promptguard/promptguard/releases/latest/download/pg-linux-amd64 -o pg
chmod +x pg

# Windows
curl -L https://github.com/promptguard/promptguard/releases/latest/download/pg-windows-amd64.exe -o pg.exe
```

**Option 2: Install from Source**
```bash
git clone https://github.com/promptguard/promptguard.git
cd promptguard
go build -o pg main.go
```

### Basic Usage

1. **Create a configuration file** (`promptguard.yaml`):
```yaml
description: "My LLM prompt tests"
prompts:
  - prompts/*.prompt
providers:
  - id: openai:gpt-4o
    config:
      temperature: 0
tests:
  - vars: {customer: "Alice", product: "Pro Plan"}
    assert:
      - type: answer-relevance
        value: "Mention Pro Plan benefits"
        threshold: 0.7
      - type: cost
        threshold: 0.003
```

2. **Create a prompt file** (`prompts/onboard.prompt`):
```
---
title: "Customer Onboarding"
---

Welcome {{.customer}} to {{.product}}! 

Please provide:
1. Getting started guide
2. Key features overview
3. Next steps

Format as JSON with: welcome_message, features, next_steps
```

3. **Set your API key**:
```bash
export OPENAI_API_KEY="sk-..."
```

4. **Run tests**:
```bash
./pg test
```

## 📋 Features

### ✅ Must-Have Features (v0.1)
- **Multiple Providers**: OpenAI, Anthropic, Mistral support
- **Rich Assertions**: Cost thresholds, relevance scoring, JSON validation
- **CLI Commands**: `pg test` (local), `pg ci` (CI/CD), `pg view` (interactive)
- **Configuration**: YAML-based with variable substitution
- **Reporting**: JSON, JUnit XML, HTML, Markdown formats

### 🎯 Assertion Types
- **`answer-relevance`**: Semantic similarity scoring
- **`contains-json`**: JSON structure validation with schema
- **`cost`**: Token cost threshold enforcement
- **`llm-rubric`**: LLM-graded quality assessment
- **`toxicity`**: Content safety detection
- **`jailbreak`**: Prompt injection detection

### 📊 CI/CD Integration
- **GitHub Actions**: Ready-to-use action with annotations
- **Baseline Comparison**: Detect regressions automatically
- **Artifacts**: HTML reports, metrics, and diffs
- **Badge Generation**: Show test status in README

## 🔧 CLI Commands

### `pg test` - Local Testing
```bash
pg test [flags]

Flags:
  -o, --output string        Output format (console, json, junit, html)
      --output-file string   Output file path
  -p, --parallel int         Parallel executions (default 1)
      --update-baseline      Update baseline results
      --filter strings       Filter tests by pattern
```

### `pg ci` - CI/CD Mode
```bash
pg ci [flags]

Flags:
      --baseline-path string    Baseline results path (default ".promptguard/baseline.json")
      --artifacts-dir string    Artifacts directory (default "artifacts")
      --github-annotations      Generate GitHub annotations (default true)
      --commit-sha string       Git commit SHA
      --pr-number string        Pull request number
```

### `pg view` - Interactive Viewer
```bash
pg view [flags]

Flags:
  -p, --port int              Server port (default 8080)
      --results-file string   Results file path (default "artifacts/results.json")
      --open-browser          Auto-open browser (default true)
```

## 📁 Project Structure

```
my-project/
├── promptguard.yaml           # Main configuration
├── prompts/                   # Prompt templates
│   ├── onboard.prompt
│   ├── invoice.prompt
│   └── newsletter.prompt
├── .promptguard/             # PromptGaurd by Chandresh data
│   ├── baseline.json         # Baseline results
│   └── metrics.db           # Historical metrics
├── artifacts/               # Generated reports
│   ├── results.json
│   ├── promptguard.html
│   └── junit.xml
└── .github/
    └── workflows/
        └── promptguard.yml  # CI workflow
```

## ⚙️ Configuration Reference

### Complete Configuration Example
```yaml
description: "E-commerce prompt tests"

# Prompt files (supports glob patterns)
prompts:
  - prompts/onboard.prompt
  - prompts/**/*.prompt

# LLM providers
providers:
  - id: openai:gpt-4o
    config:
      temperature: 0
      max_tokens: 1000
  
  - id: anthropic:claude-3-haiku
    config:
      temperature: 0.2

# Test cases
tests:
  - name: "onboard-pro-user"
    vars:
      customer: "Alice Johnson"
      product: "Pro Plan"
      features: ["Analytics", "API", "Support"]
    assert:
      - type: answer-relevance
        value: "Mention Pro Plan upgrade benefits"
        threshold: 0.8
      - type: contains-json
        value:
          type: object
          required: ["welcome_message", "next_steps"]
      - type: cost
        threshold: 0.003

# Global settings
settings:
  costBudget: 0.05      # Total budget per run
  timeout: 30           # Request timeout (seconds)
  maxRetries: 2         # Retry failed requests
  cacheResults: true    # Cache responses
```

### Prompt Template Format
```markdown
---
title: "Prompt Title"
description: "What this prompt does"
version: "1.0"
tags: ["onboarding", "json"]
---

You are a helpful assistant for {{.customer}}.

Task: Create onboarding content for {{.product}}.

Requirements:
1. Personal greeting
2. Feature overview: {{range .features}}{{.}}, {{end}}
3. Next steps

Response format: JSON with welcome_message and next_steps fields.
```

## 🎭 GitHub Actions Integration

### Basic Workflow
```yaml
name: PromptGuard Tests

on: [push, pull_request]

jobs:
  test-prompts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run PromptGuard
        uses: promptguard/run@v1
        with:
          config-file: promptguard.yaml
          openai-api-key: ${{ secrets.OPENAI_API_KEY }}
          fail-on-regression: true
```

### Advanced Workflow with Multiple Providers
```yaml
- name: Run PromptGuard
  uses: promptguard/run@v1
  with:
    config-file: promptguard.yaml
    openai-api-key: ${{ secrets.OPENAI_API_KEY }}
    anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}
    baseline-path: .promptguard/baseline.json
    artifacts-dir: test-results
```

## 📊 Example Output

### Console Output
```
=== PromptGuard Test Results ===
Generated: 2024-12-21T10:30:00Z

Summary:
  Tests: 4
  Passed: 3
  Failed: 1
  Cost: $0.0234
  Duration: 2.3s

Failures:
  ❌ invoice-generation
     contains-json: Required field missing: total_due
     cost: $0.0045 exceeds threshold $0.003
```

### HTML Report Features
- 🎯 Interactive test result explorer
- 📊 Cost and performance metrics
- 🔍 Side-by-side diff viewer
- 📈 Historical trend charts
- 🎮 "What-if" scenario testing

## 🛠️ Development

### Prerequisites
- Go 1.21+
- OpenAI/Anthropic/Mistral API keys

### Build from Source
```bash
git clone https://github.com/promptguard/promptguard.git
cd promptguard
go mod download
go build -o pg main.go
```

### Run Tests
```bash
go test ./...
```

### Example Configuration
The repository includes a complete example in the root directory. Set your API key and run:

```bash
export OPENAI_API_KEY="sk-..."
./pg test
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Key Areas for Contribution
- 🔌 Additional LLM providers (Ollama, Azure OpenAI, etc.)
- 🧪 New assertion types (semantic similarity, fact checking)
- 📱 UI/UX improvements for the web viewer
- 📊 Advanced analytics and reporting
- 🔧 IDE integrations and extensions

## 🗺️ Roadmap

### v0.1 - Core MVP ✅
- [x] CLI with test/ci/view commands
- [x] OpenAI provider integration
- [x] Basic assertions (cost, relevance, JSON)
- [x] HTML/JSON/JUnit reporting

### v0.2 - CI/CD Focus
- [ ] GitHub Action marketplace release
- [ ] Baseline comparison and regression detection
- [ ] Enhanced diff viewer with syntax highlighting
- [ ] Slack/Teams notifications

### v0.3 - Advanced Features
- [ ] Anthropic and Mistral provider support
- [ ] LLM-graded assertions (rubric scoring)
- [ ] Toxicity and jailbreak detection
- [ ] Interactive prompt playground

### v1.0 - Enterprise Ready
- [ ] Plugin SDK for custom assertions
- [ ] Enterprise authentication (SSO, RBAC)
- [ ] Advanced analytics dashboard
- [ ] Multi-tenant support

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Inspired by the excellent work of:
- [Promptfoo](https://github.com/promptfoo/promptfoo) - LLM evaluation framework
- [PromptLayer](https://promptlayer.com/) - LLM observability platform
- [LangChain](https://github.com/langchain-ai/langchain) - LLM application framework

---

**Made with ❤️ by the PromptGuard team**

[📖 Documentation](https://promptguard.dev) • [💬 Discord](https://discord.gg/promptguard) • [🐛 Issues](https://github.com/promptguard/promptguard/issues) • [🚀 Releases](https://github.com/promptguard/promptguard/releases)
