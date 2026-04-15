package cmd

import (
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
