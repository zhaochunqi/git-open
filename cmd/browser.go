package cmd

import (
	"errors"

	"github.com/pkg/browser"
)

// ErrMockBrowser is used for testing browser errors
var ErrMockBrowser = errors.New("mock browser error")

// OpenURLInBrowser is exported for testing
var OpenURLInBrowser = browser.OpenURL

func openURLInBrowserFunc(url string) error {
	return OpenURLInBrowser(url)
}
