# Git Open

[![Go Report Card](https://goreportcard.com/badge/github.com/zhaochunqi/git-open)](https://goreportcard.com/report/github.com/zhaochunqi/git-open)

This is a Go-based tool that allows you to open the current repository in a web browser. It's a simple yet efficient solution for quickly accessing your project's online resources.

## Features

* Easy Access: Open your repository in a web browser with a single command.
* Cross-Platform: Built using Go, the tool is compatible with various operating systems.
* Lightweight: Minimal dependencies and efficient code ensure the tool runs smoothly.

## Installation

There are multiple ways to install the tool:

### Using Go

Ensure you have Go installed on your system. Then, use the following command:

```sh
go install github.com/zhaochunqi/git-open@latest
```

### Using mise with ubi

If you prefer using mise, you can install the tool with:

```sh
mise use -g ubi:zhaochunqi/git-open
```

This method allows for easy version management and isolation.

## Usage

Navigate to your project's directory and run the following command:

`git-open` or `git open`

This will open your repository in the default web browser.

## Contributing

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please submit a pull request. For major changes, please open an issue first to discuss what you would like to change.

