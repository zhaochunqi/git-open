package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhaochunqi/git-open/internal/testhelper"
)

// Test_repoCmd_ChdirFlag verifies that -C <path> runs as if started in <path>,
// using the git repository there instead of the current working directory.
func Test_repoCmd_ChdirFlag(t *testing.T) {
	// Set up a test repo in a separate temp directory.
	repoDir, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	// Change to a directory that is not a git repository, then resolve the
	// repo from repoDir via the -C flag.
	nonRepoDir, err := os.MkdirTemp("", "git-open-nonrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(nonRepoDir)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nonRepoDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	oldArgs := os.Args
	os.Args = []string{"git-open", "-C", repoDir, "repo"}
	t.Cleanup(func() { os.Args = oldArgs })
	t.Cleanup(func() { chdirPaths = nil })

	if err := Execute(); err != nil {
		t.Fatalf("Execute() with -C = %v", err)
	}
}

// Test_Execute_UnknownFlagHint verifies that flag mistakes hint at --help
// of the offending command instead of dumping the full usage.
func Test_Execute_UnknownFlagHint(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantSub string
	}{
		{
			name:    "root command flag",
			args:    []string{"git-open", "--bogus"},
			wantSub: "unknown flag: --bogus\nRun 'git-open --help' for usage.",
		},
		{
			name:    "subcommand flag",
			args:    []string{"git-open", "repo", "--bogus"},
			wantSub: "unknown flag: --bogus\nRun 'git-open repo --help' for usage.",
		},
		{
			name:    "unknown command",
			args:    []string{"git-open", "bogus"},
			wantSub: "unknown command \"bogus\" for \"git-open\"\nRun 'git-open --help' for usage.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			os.Args = tt.args
			t.Cleanup(func() { os.Args = oldArgs })

			err := Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("Execute() error = %q, want message containing %q", err.Error(), tt.wantSub)
			}
		})
	}
}

// Test_repoCmd_ChdirFlag_InvalidPath verifies a non-existent -C path errors out.
func Test_repoCmd_ChdirFlag_InvalidPath(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"git-open", "-C", filepath.Join(t.TempDir(), "does-not-exist"), "repo"}
	t.Cleanup(func() { os.Args = oldArgs })
	t.Cleanup(func() { chdirPaths = nil })

	err := Execute()
	if err == nil {
		t.Fatal("Execute() expected error for invalid -C path, got nil")
	}
	if !strings.Contains(err.Error(), "changing directory") {
		t.Fatalf("Execute() error = %v, want message containing 'changing directory'", err)
	}
}
