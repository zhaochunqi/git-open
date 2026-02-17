package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_getCurrentGitDirectory(t *testing.T) {
	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	tests := []struct {
		name    string
		setup   func()
		want    bool
		wantErr bool
	}{
		{
			name:    "valid git repo",
			setup:   func() {},
			want:    true,
			wantErr: false,
		},
		{
			name: "non-git directory",
			setup: func() {
				tmpDir, err := os.MkdirTemp("", "non-git")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatal(err)
				}
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := getCurrentGitDirectory()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentGitDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want && got == nil {
				t.Error("getCurrentGitDirectory() = nil, want valid repository")
			}
		})
	}
}

func Test_getCurrentGitDirectory_Worktree(t *testing.T) {
	_, cleanup := testhelper.SetupTestWorktree(t, "https://github.com/zhaochunqi/git-open.git", "feature-branch")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error in worktree = %v", err)
	}
	if repo == nil {
		t.Error("getCurrentGitDirectory() = nil, want valid repository")
	}
}

func Test_getRemoteURL_Worktree(t *testing.T) {
	_, cleanup := testhelper.SetupTestWorktree(t, "https://github.com/zhaochunqi/git-open.git", "feature-branch")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error = %v", err)
	}

	got, err := getRemoteURL(repo)
	if err != nil {
		t.Errorf("getRemoteURL() error = %v", err)
		return
	}
	expectedURL := "https://github.com/zhaochunqi/git-open.git"
	if got != expectedURL {
		t.Errorf("getRemoteURL() = %v, want %v", got, expectedURL)
	}
}

func Test_getBranchName_Worktree(t *testing.T) {
	_, cleanup := testhelper.SetupTestWorktree(t, "https://github.com/zhaochunqi/git-open.git", "feature-branch")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error = %v", err)
	}

	got, err := getBranchName(repo)
	if err != nil {
		t.Errorf("getBranchName() error = %v", err)
		return
	}
	expectedBranch := "feature-branch"
	if got != expectedBranch {
		t.Errorf("getBranchName() = %v, want %v", got, expectedBranch)
	}
}

// Helper function to extract the core logic from getRemoteURL for testing
func getRemoteURLFromConfig(cfg *config.RemoteConfig) (string, error) {
	urls := cfg.URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("remote URL not found")
	}

	return urls[0], nil
}

func Test_getRemoteURL(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		setup      func(t *testing.T, repo *git.Repository)
		customTest func(t *testing.T, repo *git.Repository) // For custom test logic
		want       string
		wantErr    bool
	}{
		{
			name:      "github https url",
			remoteURL: "https://github.com/zhaochunqi/git-open.git",
			want:      "https://github.com/zhaochunqi/git-open.git",
			wantErr:   false,
		},
		{
			name:      "github ssh url",
			remoteURL: "git@github.com:zhaochunqi/git-open.git",
			want:      "git@github.com:zhaochunqi/git-open.git",
			wantErr:   false,
		},
		{
			name:      "gitlab https url",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      "https://gitlab.com/user/repo.git",
			wantErr:   false,
		},
		{
			name:      "no remote url",
			remoteURL: "",
			want:      "",
			wantErr:   true,
		},
		{
			name:      "empty remote urls",
			remoteURL: "https://github.com/zhaochunqi/git-open.git",
			setup: func(t *testing.T, repo *git.Repository) {
				// Remove all remotes
				cfg, err := repo.Config()
				if err != nil {
					t.Fatal(err)
				}
				cfg.Remotes = make(map[string]*config.RemoteConfig)
				err = repo.SetConfig(cfg)
				if err != nil {
					t.Fatal(err)
				}
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If we have a custom test, run it instead of the standard test
			if tt.customTest != nil {
				_, cleanup := testhelper.SetupTestRepo(t, tt.remoteURL, "main")
				defer cleanup()

				repo, err := getCurrentGitDirectory()
				if err != nil {
					t.Fatal(err)
				}

				if tt.setup != nil {
					tt.setup(t, repo)
				}

				// Run custom test logic
				tt.customTest(t, repo)
				return
			}

			// Standard test path
			_, cleanup := testhelper.SetupTestRepo(t, tt.remoteURL, "main")
			defer cleanup()

			repo, err := getCurrentGitDirectory()
			if err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				tt.setup(t, repo)
			}

			got, err := getRemoteURL(repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRemoteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertToWebURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "github https url",
			url:     "https://github.com/zhaochunqi/git-open.git",
			want:    "https://github.com/zhaochunqi/git-open",
			wantErr: false,
		},
		{
			name:    "github ssh url",
			url:     "git@github.com:zhaochunqi/git-open.git",
			want:    "https://github.com/zhaochunqi/git-open",
			wantErr: false,
		},
		{
			name:    "gitlab https url",
			url:     "https://gitlab.com/user/repo.git",
			want:    "https://gitlab.com/user/repo",
			wantErr: false,
		},
		{
			name:    "gitlab ssh url",
			url:     "git@gitlab.com:user/repo.git",
			want:    "https://gitlab.com/user/repo",
			wantErr: false,
		},
		{
			name:    "invalid url",
			url:     "invalid-url",
			want:    "",
			wantErr: true,
		},
		{
			name:    "ssh url with ssh prefix",
			url:     "ssh://git@github.com/user/repo.git",
			want:    "https://github.com/user/repo",
			wantErr: false,
		},
		{
			name:    "http url without git suffix",
			url:     "http://github.com/user/repo",
			want:    "http://github.com/user/repo",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToWebURL(tt.url)
			if got != tt.want {
				t.Errorf("convertToWebURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBranchName(t *testing.T) {
	tests := []struct {
		name       string
		remoteURL  string
		branchName string
		setup      func(t *testing.T, repo *git.Repository)
		wantErr    bool
	}{
		{
			name:       "main branch",
			remoteURL:  "https://github.com/zhaochunqi/git-open.git",
			branchName: "main",
			wantErr:    false,
		},
		{
			name:       "feature branch",
			remoteURL:  "https://github.com/zhaochunqi/git-open.git",
			branchName: "feature-branch",
			wantErr:    false,
		},
		{
			name:       "uninitialized repo with HEAD reference",
			remoteURL:  "https://github.com/zhaochunqi/git-open.git",
			branchName: "main",
			wantErr:    true, // Expect error when there's no commit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.name == "uninitialized repo with HEAD reference" {
				_, cleanup = testhelper.SetupTestRepoWithoutCommit(t, tt.remoteURL, tt.branchName)
			} else {
				_, cleanup = testhelper.SetupTestRepo(t, tt.remoteURL, tt.branchName)
			}
			defer cleanup()

			repo, err := getCurrentGitDirectory()
			if err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				tt.setup(t, repo)
			}

			got, err := getBranchName(repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getBranchName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.branchName {
				t.Errorf("getBranchName() = %v, want %v", got, tt.branchName)
			}
		})
	}
}

func Test_getHostingService(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      HostingService
	}{
		{
			name:      "github",
			remoteURL: "https://github.com/user/repo.git",
			want:      GitHub,
		},
		{
			name:      "gitlab",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      GitLab,
		},
		{
			name:      "bitbucket",
			remoteURL: "https://bitbucket.org/user/repo.git",
			want:      Bitbucket,
		},
		{
			name:      "unknown service",
			remoteURL: "https://example.com/user/repo.git",
			want:      Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHostingService(tt.remoteURL); got != tt.want {
				t.Errorf("getHostingService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildBranchURL(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		branch    string
		remoteURL string
		want      string
	}{
		{
			name:      "github branch url",
			baseURL:   "https://github.com/user/repo",
			branch:    "feature",
			remoteURL: "https://github.com/user/repo.git",
			want:      "https://github.com/user/repo/tree/feature",
		},
		{
			name:      "gitlab branch url",
			baseURL:   "https://gitlab.com/user/repo",
			branch:    "feature",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      "https://gitlab.com/user/repo/-/tree/feature",
		},
		{
			name:      "bitbucket branch url",
			baseURL:   "https://bitbucket.org/user/repo",
			branch:    "feature",
			remoteURL: "https://bitbucket.org/user/repo.git",
			want:      "https://bitbucket.org/user/repo/src/feature",
		},
		{
			name:      "unknown service defaults to github style",
			baseURL:   "https://example.com/user/repo",
			branch:    "feature",
			remoteURL: "https://example.com/user/repo.git",
			want:      "https://example.com/user/repo/tree/feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildBranchURL(tt.baseURL, tt.branch, tt.remoteURL); got != tt.want {
				t.Errorf("buildBranchURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkConvertToWebURL benchmarks the URL conversion function
func BenchmarkConvertToWebURL(b *testing.B) {
	urls := []string{
		"https://github.com/zhaochunqi/git-open.git",
		"git@github.com:zhaochunqi/git-open.git",
		"https://gitlab.com/user/repo.git",
		"git@gitlab.com:user/repo.git",
	}

	for _, url := range urls {
		b.Run(url, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = convertToWebURL(url)
			}
		})
	}
}
func Test_getBranchName_Error(t *testing.T) {
	// Save original function
	originalGetBranchNameFunc := getBranchNameFunc
	defer func() {
		getBranchNameFunc = originalGetBranchNameFunc
	}()

	// Mock getBranchNameFunc to return an error
	getBranchNameFunc = func(repo *git.Repository) (string, error) {
		return "", fmt.Errorf("error getting HEAD: mock error")
	}

	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatal(err)
	}

	_, err = getBranchName(repo)
	if err == nil {
		t.Error("Expected error from getBranchName, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "error getting HEAD") {
		t.Errorf("Expected error message to contain 'error getting HEAD', got '%s'", err.Error())
	}
}

func Test_getRemoteURL_EmptyURLs(t *testing.T) {
	// Save original function
	originalGetRemoteURLFunc := getRemoteURLFunc
	defer func() {
		getRemoteURLFunc = originalGetRemoteURLFunc
	}()

	// Mock getRemoteURLFunc to simulate empty URLs error
	getRemoteURLFunc = func(repo *git.Repository) (string, error) {
		return "", fmt.Errorf("remote URL not found")
	}

	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatal(err)
	}

	url, err := getRemoteURL(repo)
	if err == nil {
		t.Error("Expected error for empty URLs, got nil")
	}
	if url != "" {
		t.Errorf("Expected empty URL, got %s", url)
	}
	if err != nil && err.Error() != "remote URL not found" {
		t.Errorf("Expected error message 'remote URL not found', got '%s'", err.Error())
	}
}
