package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/zhaochunqi/git-open/cmd"
)

func Test_main(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	// Save original OpenURLInBrowser function and restore after test
	originalOpenURL := cmd.OpenURLInBrowser
	defer func() {
		cmd.OpenURLInBrowser = originalOpenURL
	}()

	// Mock OpenURLInBrowser function
	openedURL := ""
	cmd.OpenURLInBrowser = func(url string) error {
		openedURL = url
		return nil
	}

	tests := []struct {
		name      string
		args      []string
		wantURL   string
		wantError bool
	}{
		{
			name:      "default behavior",
			args:      []string{"git-open"},
			wantURL:   "https://github.com/zhaochunqi/git-open/tree/feat/open-branch",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test repository for this test case
			_, cleanup := cmd.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "feat/open-branch")
			t.Cleanup(cleanup)

			os.Args = tt.args
			openedURL = ""

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			main()

			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.wantError && output == "" {
				t.Errorf("Expected error output, but got none")
			}

			if openedURL != tt.wantURL {
				t.Errorf("OpenURLInBrowser called with %v, want %v", openedURL, tt.wantURL)
			}
		})
	}
}