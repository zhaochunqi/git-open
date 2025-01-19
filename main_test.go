package main

import (
	"testing"
	"os"

	"github.com/zhaochunqi/git-open/cmd"
)

func Test_main(t *testing.T) {
	// Save original function
	original := cmd.OpenURLInBrowser
	defer func() {
		cmd.OpenURLInBrowser = original
	}()

	// Mock the browser function
	cmd.OpenURLInBrowser = func(url string) error {
		return nil
	}

	// Set the plain flag to avoid opening browser
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()
	os.Args = []string{"git-open", "--plain"}

	// Run main
	main()
}
