package testhelper

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// SetupTestRepo creates a temporary git repository for testing.
// It returns the temporary directory path and a cleanup function.
func SetupTestRepo(t *testing.T, remoteURL string, branchName string) (string, func()) {
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

	// Create a worktree and add a file
	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Add("test.txt")
	if err != nil {
		t.Fatal(err)
	}

	// Create an initial commit
	_, err = w.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create and checkout the specified branch if provided
	if branchName != "" {
		headRef, err := repo.Head()
		if err != nil {
			t.Fatal(err)
		}
		branchRef := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+branchName), headRef.Hash())
		err = repo.Storer.SetReference(branchRef)
		if err != nil {
			t.Fatal(err)
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + branchName),
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
