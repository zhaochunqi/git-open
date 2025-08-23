package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/zhaochunqi/git-open/cmd"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_runApp(t *testing.T) {
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

	// Save original cmd.Execute and restore after test
	originalExecute := cmd.Execute
	defer func() {
		cmd.Execute = originalExecute
	}()

	// Save original exitFunc and restore after test
	originalExitFunc := exitFunc
	defer func() {
		exitFunc = originalExitFunc
	}()

	// Mock OpenURLInBrowser function
	var openedURL string
	cmd.OpenURLInBrowser = func(url string) error {
		openedURL = url
		return nil
	}

	tests := []struct {
		name      string
		args      []string
		wantURL   string
		executeErr error // New field to simulate cmd.Execute error
		wantErr   bool  // Expect runApp to return an error
	}{
		{
			name:      "default behavior",
			args:      []string{"git-open"},
			wantURL:   "https://github.com/zhaochunqi/git-open/tree/feat/open-branch",
			executeErr: nil,
			wantErr:   false,
		},
		{
			name:      "error from cmd.Execute",
			args:      []string{"git-open"},
			wantURL:   "", // Not relevant for this test
			executeErr: errors.New("mock execute error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		// Mock os.Exit to capture the exit code
		var actualExitCode int
		exitFunc = func(code int) {
			actualExitCode = code
			// We don't want to actually exit during tests
			panic("os.Exit called") 
		}

		t.Run(tt.name, func(t *testing.T) {
			// Setup test repository for this test case
			_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "feat/open-branch")
			t.Cleanup(cleanup)

			os.Args = tt.args
			openedURL = ""

			// Mock cmd.Execute to return a specific error or call original
			cmd.Execute = func() error {
				if tt.executeErr != nil {
					return tt.executeErr
				}
				// Call the original cmd.Execute if no error is simulated
				return originalExecute()
			}

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Call runApp and recover from panic if os.Exit is called
			func() {
				defer func() {
					if r := recover(); r != nil && r.(string) != "os.Exit called" {
						t.Fatalf("Unexpected panic: %v", r)
					}
				}()
				if err := runApp(); (err != nil) != tt.wantErr {
					t.Errorf("runApp() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()

			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.wantErr && output == "" {
				t.Errorf("Expected error output, but got none")
			}

			if tt.executeErr == nil && openedURL != tt.wantURL {
				t.Errorf("OpenURLInBrowser called with %v, want %v", openedURL, tt.wantURL)
			}

			// For error case, check if os.Exit was called with 1
			if tt.wantErr && actualExitCode != 1 {
				t.Errorf("Expected os.Exit(1) to be called, but got %v", actualExitCode)
			}
		})
	}
}

func Test_main_func(t *testing.T) {
	// This test ensures that the main function calls runApp
	// and handles os.Exit correctly.

	// Save original exitFunc and restore after test
	originalExitFunc := exitFunc
	defer func() {
		exitFunc = originalExitFunc
	}()

	// Save original cmd.Execute and restore after test
	originalExecute := cmd.Execute
	defer func() {
		cmd.Execute = originalExecute
	}()

	// Mock os.Exit to capture the exit code
	var actualExitCode int
	exitFunc = func(code int) {
		actualExitCode = code
		panic("os.Exit called") // Panic to stop execution
	}

	// Mock cmd.Execute to return an error
	cmd.Execute = func() error {
		return errors.New("mock error for main func test")
	}

	// Call main and recover from panic
	func() {
		defer func() {
			if r := recover(); r != nil && r.(string) != "os.Exit called" {
				t.Fatalf("Unexpected panic: %v", r)
			}
		}()
		main()
	}()

	// Verify that os.Exit was called with 1
	if actualExitCode != 1 {
		t.Errorf("Expected os.Exit(1) to be called, but got %v", actualExitCode)
	}
}