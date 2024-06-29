package cmd

import (
	"reflect"
	"testing"

	"github.com/go-git/go-git/v5"
)

func Test_getCurrentGitDirectory(t *testing.T) {
	tests := []struct {
		name    string
		want    *git.Repository
		wantErr bool
	}{
		// TODO: Add test cases.

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCurrentGitDirectory()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCurrentGitDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCurrentGitDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRemoteURL(t *testing.T) {
	type args struct {
		repo *git.Repository
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRemoteURL(tt.args.repo)
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
