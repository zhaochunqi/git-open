package cmd

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func Test_getCurrentGitDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize test git repository
	_, err = git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Save current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Change to test directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	tests := []struct {
		name    string
		want    bool
		wantErr bool
	}{
		{
			name:    "valid git repo",
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize test git repository
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	// Add remote repository
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/zhaochunqi/git-open.git"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		repo    *git.Repository
		want    string
		wantErr bool
	}{
		{
			name:    "valid remote",
			repo:    repo,
			want:    "https://github.com/zhaochunqi/git-open.git",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRemoteURL(tt.repo)
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
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "github_ssh",
			args: args{
				url: "ssh://git@github.com/zhaochunqi/git-open.git",
			},
			want: "https://github.com/zhaochunqi/git-open",
		},
		{
			name: "github_https",
			args: args{
				url: "https://github.com/zhaochunqi/blog.git",
			},
			want: "https://github.com/zhaochunqi/blog",
		},
		{
			name: "gitlab_ssh",
			args: args{
				url: "git@gitlab.com:gitlab-org/govern/security-policies/alexander-test-group/alexander-test-subgroup/sub-sub-gitlab-org/sub-sub-group-project.git",
			},
			want: "https://gitlab.com/gitlab-org/govern/security-policies/alexander-test-group/alexander-test-subgroup/sub-sub-gitlab-org/sub-sub-group-project",
		},
		{
			name: "gitlab_https",
			args: args{
				url: "https://gitlab.com/gitlab-org/govern/security-policies/alexander-test-group/alexander-test-subgroup/sub-sub-gitlab-org/sub-sub-group-project.git",
			},
			want: "https://gitlab.com/gitlab-org/govern/security-policies/alexander-test-group/alexander-test-subgroup/sub-sub-gitlab-org/sub-sub-group-project",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToWebURL(tt.args.url); got != tt.want {
				t.Errorf("convertToWebURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
