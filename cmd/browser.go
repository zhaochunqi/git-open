package cmd

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// ErrMockBrowser is used for testing browser errors
var ErrMockBrowser = errors.New("mock browser error")

// OpenURLInBrowser is exported for testing
var OpenURLInBrowser = openURLInBrowser

// getPlatform returns the current platform, can be mocked for testing
var getPlatform = func() string {
	return runtime.GOOS
}

var commandRunner = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

var BrowserCommand string

func openURLInBrowser(url string) error {
	platform := getPlatform()

	customBrowser := strings.TrimSpace(BrowserCommand)
	if customBrowser != "" {
		return commandRunner(customBrowser, url)
	}

	// On Linux, use xdg-open with output redirection to suppress messages
	if platform == "linux" {
		return openWithXdgOpen(url)
	}
	// On macOS, use open command with output redirection to suppress messages
	if platform == "darwin" {
		return openWithMacOSOpen(url)
	}
	// On Windows, use start command with output redirection to suppress messages
	if platform == "windows" {
		return openWithWindowsStart(url)
	}
	// For other platforms, return an error
	return errors.New("unsupported platform: " + platform)
}

func openWithXdgOpen(url string) error {
	return commandRunner("xdg-open", url)
}

func openWithMacOSOpen(url string) error {
	return commandRunner("open", url)
}

func openWithWindowsStart(url string) error {
	return commandRunner("cmd", "/c", "start", "", url)
}

func openURLInBrowserFunc(url string) error {
	return OpenURLInBrowser(url)
}
