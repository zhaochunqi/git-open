package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_rootCmd_InvalidRemoteURLFormat(t *testing.T) {
	originalGetRemoteURLFunc := getRemoteURLFunc
	defer func() { getRemoteURLFunc = originalGetRemoteURLFunc }()

	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
	defer cleanup()

	getRemoteURLFunc = func(repo *git.Repository) (string, error) {
		return "invalid-remote", nil
	}

	cmd := &cobra.Command{}
	cmd.SetOut(new(bytes.Buffer))
	cmd.Flags().Bool("plain", false, "")
	cmd.Flags().Bool("version", false, "")

	err := rootCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("rootCmd.RunE() expected error for invalid remote URL format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported remote URL format") {
		t.Fatalf("rootCmd.RunE() error = %v, want message containing 'unsupported remote URL format'", err)
	}
}
