package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_repoNameFromWebURL(t *testing.T) {
	tests := []struct {
		name   string
		webURL string
		want   string
	}{
		{"https URL", "https://github.com/zhaochunqi/git-open", "github.com/zhaochunqi/git-open"},
		{"http URL", "http://gitlab.com/user/repo", "gitlab.com/user/repo"},
		{"trailing slash", "https://github.com/user/repo/", "github.com/user/repo"},
		{"no scheme", "github.com/user/repo", "github.com/user/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repoNameFromWebURL(tt.webURL); got != tt.want {
				t.Errorf("repoNameFromWebURL(%q) = %q, want %q", tt.webURL, got, tt.want)
			}
		})
	}
}

func Test_repoCmd(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      string
	}{
		{"https URL", "https://github.com/zhaochunqi/git-open.git", "github.com/zhaochunqi/git-open\n"},
		{"ssh URL", "git@github.com:zhaochunqi/git-open.git", "github.com/zhaochunqi/git-open\n"},
		{"gitlab URL", "https://gitlab.com/user/repo.git", "gitlab.com/user/repo\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := testhelper.SetupTestRepo(t, tt.remoteURL, "main")
			defer cleanup()

			buf := new(bytes.Buffer)
			cmd := &cobra.Command{}
			cmd.SetOut(buf)

			if err := repoCmd.RunE(cmd, []string{}); err != nil {
				t.Fatalf("repoCmd.RunE() error = %v", err)
			}

			if got := buf.String(); got != tt.want {
				t.Errorf("repoCmd output = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_repoCmd_InvalidRemoteURLFormat(t *testing.T) {
	originalGetRemoteURLFunc := getRemoteURLFunc
	defer func() { getRemoteURLFunc = originalGetRemoteURLFunc }()

	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
	defer cleanup()

	getRemoteURLFunc = func(repo *git.Repository) (string, error) {
		return "invalid-remote", nil
	}

	cmd := &cobra.Command{}
	cmd.SetOut(new(bytes.Buffer))

	err := repoCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("repoCmd.RunE() expected error for invalid remote URL format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported remote URL format") {
		t.Fatalf("repoCmd.RunE() error = %v, want message containing 'unsupported remote URL format'", err)
	}
}
