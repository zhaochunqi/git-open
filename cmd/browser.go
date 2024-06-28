package cmd

import "github.com/pkg/browser"

func openURLInBrowser(url string) error {
	return browser.OpenURL(url)
}
