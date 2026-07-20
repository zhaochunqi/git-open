# Git Open

[![codecov](https://codecov.io/github/zhaochunqi/git-open/graph/badge.svg?token=TXC9ZOSHFT)](https://codecov.io/github/zhaochunqi/git-open)

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
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Darwin_x86_64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Darwin_arm64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/

# Linux (x64)
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Linux_x86_64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/
```


### Nix

Using flakes (recommended):

```sh
# one-shot run without installing
nix run github:zhaochunqi/git-open

# install into your profile
nix profile install github:zhaochunqi/git-open

# or pin a release tag
nix profile install github:zhaochunqi/git-open/v2.4.2
```

From a local checkout:

```sh
nix build
./result/bin/git-open version

# development shell with Go toolchain
nix develop
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

To print the repository name (e.g. `github.com/zhaochunqi/git-open`):

`git-open repo`

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
