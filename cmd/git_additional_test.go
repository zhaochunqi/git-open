package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_getBranchName_DetachedHEAD(t *testing.T) {
	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error = %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("repo.Head() error = %v", err)
	}

	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.HEAD, head.Hash())); err != nil {
		t.Fatalf("set detached HEAD failed: %v", err)
	}

	branch, err := getBranchName(repo)
	if err == nil {
		t.Fatalf("getBranchName() expected error on detached HEAD, got branch=%q", branch)
	}
	if !strings.Contains(err.Error(), "error getting HEAD") {
		t.Fatalf("getBranchName() error = %v, want message containing 'error getting HEAD'", err)
	}
}

// Regression test: repositories that enable config extensions unknown to
// go-git (e.g. extensions.worktreeConfig) must still open successfully.
func Test_resolveWebURL_WorktreeConfigExtension(t *testing.T) {
	tmpDir, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	configPath := filepath.Join(tmpDir, ".git", "config")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open config failed: %v", err)
	}
	if _, err := f.WriteString("[extensions]\n\tworktreeConfig = true\n"); err != nil {
		f.Close()
		t.Fatalf("append extensions section failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close config failed: %v", err)
	}

	repo, remoteURL, webURL, err := resolveWebURL()
	if err != nil {
		t.Fatalf("resolveWebURL() error = %v", err)
	}
	if remoteURL != "https://github.com/zhaochunqi/git-open.git" {
		t.Errorf("resolveWebURL() remoteURL = %q, want %q", remoteURL, "https://github.com/zhaochunqi/git-open.git")
	}
	if webURL != "https://github.com/zhaochunqi/git-open" {
		t.Errorf("resolveWebURL() webURL = %q, want %q", webURL, "https://github.com/zhaochunqi/git-open")
	}

	branch, err := getBranchName(repo)
	if err != nil {
		t.Fatalf("getBranchName() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("getBranchName() = %q, want %q", branch, "main")
	}
}
