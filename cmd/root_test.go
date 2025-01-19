package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"
)

func setupTestRepo(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize git repository
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Add remote
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/zhaochunqi/git-open.git"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Chdir(currentDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func Test_rootCmd(t *testing.T) {
	// Setup test repository
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Save original openURLInBrowser function
	original := OpenURLInBrowser
	defer func() {
		OpenURLInBrowser = original
	}()

	// Mock openURLInBrowser function
	OpenURLInBrowser = func(url string) error {
		return nil
	}

	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{
			name:       "with plain flag",
			args:       []string{"--plain"},
			wantOutput: "Web URL: https://github.com/zhaochunqi/git-open\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for testing
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			
			// Set flags
			cmd.Flags().Bool("plain", false, "")
			if err := cmd.Flags().Set("plain", "true"); err != nil {
				t.Fatal(err)
			}

			// Run the root command function
			runE := rootCmd.RunE
			if err := runE(cmd, tt.args); err != nil {
				t.Errorf("rootCmd.RunE() error = %v", err)
				return
			}
			
			if got := buf.String(); got != tt.wantOutput {
				t.Errorf("root command output = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

func Test_Execute(t *testing.T) {
	// Setup test repository
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Save original openURLInBrowser function
	original := OpenURLInBrowser
	defer func() {
		OpenURLInBrowser = original
	}()

	// Mock openURLInBrowser function
	OpenURLInBrowser = func(url string) error {
		return nil
	}

	// Save original args
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
	}()

	// Set args for testing
	os.Args = []string{"git-open", "--plain"}

	// Save stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Restore stdout
	defer func() {
		w.Close()
		os.Stdout = oldStdout
	}()

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("rootCmd.Execute() error = %v", err)
		return
	}

	// Read output
	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Error(err)
		return
	}

	wantOutput := "Web URL: https://github.com/zhaochunqi/git-open\n"
	if got := buf.String(); got != wantOutput {
		t.Errorf("Execute() output = %q, want %q", got, wantOutput)
	}
}

func Test_initConfig(t *testing.T) {
	// Create a temporary home directory
	tmpHome, err := os.MkdirTemp("", "home")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	// Save original home
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	// Set new home
	os.Setenv("HOME", tmpHome)

	// Test initConfig
	initConfig()
}
