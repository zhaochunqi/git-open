package cmd

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

// setupTestRepo creates a temporary git repository for testing.
// It returns the temporary directory path and a cleanup function.
func setupTestRepo(t *testing.T, remoteURL string) (string, func()) {
	t.Helper()
	
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

	// Add remote if URL is provided
	if remoteURL != "" {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
		if err != nil {
			t.Fatal(err)
		}
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
