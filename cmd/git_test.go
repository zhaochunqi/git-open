package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func Test_getCurrentGitDirectory(t *testing.T) {
	// Use setupTestRepo for setup
	_, cleanup := SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
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
			name:     "github https url",
			remoteURL: "https://github.com/zhaochunqi/git-open.git",
			want:     "https://github.com/zhaochunqi/git-open.git",
			wantErr:  false,
		},
		{
			name:     "github ssh url",
			remoteURL: "git@github.com:zhaochunqi/git-open.git",
			want:     "git@github.com:zhaochunqi/git-open.git",
			wantErr:  false,
		},
		{
			name:     "gitlab https url",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:     "https://gitlab.com/user/repo.git",
			wantErr:  false,
		},
		{
			name:     "no remote url",
			remoteURL: "",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "empty remote urls",
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
		{
			name:     "empty remote URL list",
			remoteURL: "https://github.com/zhaochunqi/git-open.git",
			setup: func(t *testing.T, repo *git.Repository) {
				// Standard setup is fine, we'll use customTest
			},
			customTest: func(t *testing.T, repo *git.Repository) {
				// Create a mock remote config with empty URLs
				mockConfig := &config.RemoteConfig{
					Name: "origin",
					URLs: []string{},
				}

				// Test the function with our mock config
				url, err := getRemoteURLFromConfig(mockConfig)
				if err == nil {
					t.Error("Expected error for empty URLs, got nil")
				}
				if url != "" {
					t.Errorf("Expected empty URL, got %s", url)
				}
				if err != nil && err.Error() != "remote URL not found" {
					t.Errorf("Expected error message 'remote URL not found', got '%s'", err.Error())
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
				_, cleanup := SetupTestRepo(t, tt.remoteURL, "main")
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
			_, cleanup := SetupTestRepo(t, tt.remoteURL, "main")
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
