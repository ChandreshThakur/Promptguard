# PromptGuard

This is a PromptGuard project for testing LLM prompts in CI/CD pipelines.

## Project Structure

- `prompts/` - Contains prompt template files
- `promptguard.yaml` - Main configuration file
- `.promptguard/` - Storage for baselines and metrics
- `artifacts/` - Generated test reports and artifacts

## Coding Guidelines

- Use clear, descriptive variable names in prompt templates
- Follow the established prompt format with frontmatter metadata
- Ensure all assertions are meaningful and test real business requirements
- Keep costs reasonable - aim for under $0.005 per test case
- Write prompts that are deterministic and testable

## Best Practices

- Test prompts with multiple variable combinations
- Use both positive and negative test cases
- Include cost thresholds to prevent budget overruns
- Validate JSON outputs with proper schema assertions
- Test for relevance, accuracy, and safety (toxicity/jailbreak detection)

## CI/CD Integration

This project uses PromptGuard for continuous testing of LLM prompts. The CI pipeline:

1. Runs tests against all prompt files
2. Validates responses meet assertion criteria
3. Compares against baseline results
4. Generates comprehensive reports
5. Fails the build on regressions

For more information about PromptGuard, visit: https://github.com/promptguard/promptguard
