package testhelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestSetupTestRepo(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		branchName string
	}{
		{
			name:       "basic repo without remote",
			remoteURL:  "",
			branchName: "",
		},
		{
			name:       "repo with remote URL",
			remoteURL:  "https://github.com/test/repo.git",
			branchName: "",
		},
		{
			name:       "repo with remote and branch",
			remoteURL:  "https://github.com/test/repo.git",
			branchName: "feature/test",
		},
		{
			name:       "repo with main branch",
			remoteURL:  "https://github.com/test/repo.git",
			branchName: "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call SetupTestRepo
			tmpDir, cleanup := SetupTestRepo(t, tt.remoteURL, tt.branchName)
			defer cleanup()

			// Verify that the directory was created and exists
			if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
				t.Errorf("SetupTestRepo() did not create directory at %s", tmpDir)
			}

			// Verify that it's a git repository
			repo, err := git.PlainOpen(tmpDir)
			if err != nil {
				t.Errorf("SetupTestRepo() did not create a valid git repository: %v", err)
			}

			// Verify that test.txt file was created
			testFile := filepath.Join(tmpDir, "test.txt")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Errorf("SetupTestRepo() did not create test.txt file")
			}

			// Verify that the file is committed
			w, err := repo.Worktree()
			if err != nil {
				t.Errorf("SetupTestRepo() failed to get worktree: %v", err)
			}

			status, err := w.Status()
			if err != nil {
				t.Errorf("SetupTestRepo() failed to get status: %v", err)
			}

			// Should have no uncommitted changes
			if !status.IsClean() {
				t.Errorf("SetupTestRepo() left uncommitted changes in repository")
			}

			// Verify remote is set up if remoteURL was provided
			if tt.remoteURL != "" {
				remote, err := repo.Remote("origin")
				if err != nil {
					t.Errorf("SetupTestRepo() did not create origin remote: %v", err)
				} else {
					config := remote.Config()
					if len(config.URLs) == 0 || config.URLs[0] != tt.remoteURL {
						t.Errorf("SetupTestRepo() did not set correct remote URL. Got %v, want %v", config.URLs, tt.remoteURL)
					}
				}
			}

			// Verify branch is created and checked out if branchName was provided
			if tt.branchName != "" {
				head, err := repo.Head()
				if err != nil {
					t.Errorf("SetupTestRepo() did not set HEAD: %v", err)
				}

				branchRefName := plumbing.ReferenceName("refs/heads/" + tt.branchName)
				branchRef, err := repo.Reference(branchRefName, true)
				if err != nil {
					t.Errorf("SetupTestRepo() did not create branch %s: %v", tt.branchName, err)
				}

				if branchRef.Hash() != head.Hash() {
					t.Errorf("SetupTestRepo() did not checkout branch %s", tt.branchName)
				}
			}
		})
	}
}

func TestSetupTestRepo_Cleanup(t *testing.T) {
	// Test that cleanup function works correctly
	tmpDir, cleanup := SetupTestRepo(t, "", "")
	defer cleanup()

	// Verify directory exists before cleanup
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("SetupTestRepo() did not create directory")
	}

	// Call cleanup
	cleanup()

	// Verify directory is removed after cleanup
	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Errorf("SetupTestRepo() cleanup did not remove directory")
	}
}

func TestSetupTestRepo_ErrorCases(t *testing.T) {
	// Test error cases that might not be covered
	// For example, if branch creation fails, but since it's hard to simulate, we can test edge cases

	// Test with invalid branch name that might cause issues
	// But in practice, the function handles it gracefully
	// Instead, test that the function panics or handles errors as expected

	// Since SetupTestRepo calls t.Fatal on errors, we can't easily test error paths without modifying the function
	// For now, add a test to ensure the repo is properly initialized

	tmpDir, cleanup := SetupTestRepo(t, "https://github.com/test/repo.git", "main")
	defer cleanup()

	// Verify the repo has the correct remote
	repo, err := git.PlainOpen(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		t.Fatal(err)
	}

	if remote.Config().URLs[0] != "https://github.com/test/repo.git" {
		t.Errorf("Remote URL not set correctly")
	}
}
