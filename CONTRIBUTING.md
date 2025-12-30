# Contributing to Tempo

Contributions are welcome. This document outlines the process for contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/tempo.git
   cd tempo
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/atterpac/tempo.git
   ```

## Development Setup

### Requirements

- Go 1.21 or later
- A Temporal server for testing (local or remote)

### Building

```bash
go build -o tempo ./cmd/tempo
```

### Running Tests

```bash
go test ./...
```

### Running Locally

```bash
./tempo --address localhost:7233
```

For development without a Temporal server, mock data is available.

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/workflow-diff` - New features
- `fix/connection-timeout` - Bug fixes
- `docs/readme-update` - Documentation changes

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Run `go vet` to catch common issues
- Keep functions focused and reasonably sized

### Commit Messages

Write clear, concise commit messages:
- Use present tense ("Add feature" not "Added feature")
- First line should be under 72 characters
- Reference issues when applicable ("Fix #123")

Example:
```
Add workflow diff view

Compare two workflow executions side-by-side showing
differences in inputs, outputs, and event history.

Closes #45
```

## Pull Requests

1. Sync with upstream before creating a PR:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Push your branch:
   ```bash
   git push origin your-branch-name
   ```

3. Open a pull request with:
   - Clear title describing the change
   - Description of what and why
   - Screenshots for UI changes
   - Link to related issues

4. Address review feedback by pushing additional commits

## Reporting Issues

When reporting bugs, include:
- Tempo version (`tempo --version` when available, or commit hash)
- Go version (`go version`)
- Operating system
- Temporal server version
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or screenshots

## Feature Requests

Before submitting a feature request:
- Check existing issues to avoid duplicates
- Consider if it fits the project scope
- Provide use cases and context

## Code of Conduct

- Be respectful and constructive
- Focus on the technical merit of contributions
- Help others learn and improve

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
