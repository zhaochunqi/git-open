package cmd

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/pkg/browser"
)

// ErrMockBrowser is used for testing browser errors
var ErrMockBrowser = errors.New("mock browser error")

// OpenURLInBrowser is exported for testing
var OpenURLInBrowser = openURLInBrowser

func openURLInBrowser(url string) error {
	// On Linux, use xdg-open with output redirection to suppress messages
	if runtime.GOOS == "linux" {
		return openWithXdgOpen(url)
	}
	// For other platforms, use the default browser library
	return browser.OpenURL(url)
}

func openWithXdgOpen(url string) error {
	cmd := exec.Command("xdg-open", url)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func openURLInBrowserFunc(url string) error {
	return OpenURLInBrowser(url)
}
