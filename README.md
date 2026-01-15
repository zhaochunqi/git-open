# Git Open

[![Go Report Card](https://goreportcard.com/badge/github.com/zhaochunqi/git-open)](https://goreportcard.com/report/github.com/zhaochunqi/git-open) [![codecov](https://codecov.io/github.com/zhaochunqi/git-open/graph/badge.svg?token=TXC9ZOSHFT)](https://codecov.io/github.com/zhaochunqi/git-open)

## 🚀 Quick Install

### Recommended (macOS)
```sh
brew install --cask zhaochunqi/tap/git-open
```



---

A Go-based tool that allows you to open the current repository in a web browser with a single command. It's a simple yet efficient solution for quickly accessing your project's online resources.

## Features

* Easy Access: Open your repository in a web browser with a single command.
* Cross-Platform: Built using Go, the tool is compatible with various operating systems.
* Lightweight: Minimal dependencies and efficient code ensure the tool runs smoothly.

## Installation Options

There are multiple ways to install the tool:

### Homebrew (macOS) - Recommended

The easiest method for macOS users with automatic updates:

```sh
brew install --cask zhaochunqi/tap/git-open
```

### GitHub Releases

Download the pre-compiled binary for your platform:

```sh
# macOS (Intel)
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open-darwin-amd64 -o git-open
chmod +x git-open
sudo mv git-open /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open-darwin-arm64 -o git-open
chmod +x git-open
sudo mv git-open /usr/local/bin/

# Linux (x64)
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open-linux-amd64 -o git-open
chmod +x git-open
sudo mv git-open /usr/local/bin/
```

### mise with ubi

For users who prefer mise for version management:

```sh
mise use -g ubi:zhaochunqi/git-open
```

## Usage

Navigate to your project's directory and run the following command:

`git-open` or `git open`

This will open your repository in the default web browser.

## Testing

This project follows Go testing best practices. Here's how to run the tests:

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmark tests
go test -bench=. ./...
```

### Test Structure

The tests are organized as follows:

- Unit tests for core functionality
- Integration tests for git repository operations
- Benchmark tests for performance-critical functions

The test suite uses a custom test utility package (`internal/testutil`) that provides common testing functions and fixtures.

### Test Coverage

We aim to maintain high test coverage with:
- Multiple test cases for each function
- Edge case testing
- Error condition testing
- Performance benchmarking for critical paths

### Contributing Tests

When adding new features, please ensure:
1. Add corresponding test cases
2. Include both positive and negative test scenarios
3. Add benchmark tests for performance-sensitive functions
4. Use the provided test utilities from `internal/testutil`

## Contributing

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please submit a pull request. For major changes, please open an issue first to discuss what you would like to change.
