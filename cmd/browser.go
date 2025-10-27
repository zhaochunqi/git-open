package cmd

import (
	"errors"
	"os/exec"
	"runtime"
)

// ErrMockBrowser is used for testing browser errors
var ErrMockBrowser = errors.New("mock browser error")

// OpenURLInBrowser is exported for testing
var OpenURLInBrowser = openURLInBrowser

// getPlatform returns the current platform, can be mocked for testing
var getPlatform = func() string {
	return runtime.GOOS
}

func openURLInBrowser(url string) error {
	platform := getPlatform()
	
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
	cmd := exec.Command("xdg-open", url)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func openWithMacOSOpen(url string) error {
	cmd := exec.Command("open", url)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func openWithWindowsStart(url string) error {
	cmd := exec.Command("cmd", "/c", "start", "", url)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func openURLInBrowserFunc(url string) error {
	return OpenURLInBrowser(url)
}
