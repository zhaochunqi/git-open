package cmd

import (
	"os"
	"testing"

	"github.com/zhaochunqi/git-open/internal/testutil"
)

func Test_getCurrentGitDirectory(t *testing.T) {
	// Use testutil for setup
	_, cleanup := testutil.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git")
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

func Test_getRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		remoteURL string
		want     string
		wantErr  bool
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := testutil.SetupTestRepo(t, tt.remoteURL)
			defer cleanup()

			repo, err := getCurrentGitDirectory()
			if err != nil {
				t.Fatal(err)
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
