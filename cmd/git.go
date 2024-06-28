package cmd

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
)

func getCurrentGitDirectory() (*git.Repository, error) {
	// Open the Git repository in the current working directory
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func getRemoteURL(repo *git.Repository) (string, error) {
	// Get the remote URL of the Git repository
	remote, err := repo.Remote("origin")
	if err != nil {
		return "", err
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("remote URL not found")
	}

	return urls[0], nil
}

func convertToWebURL(url string) string {
	// Replace SSH URL scheme with HTTPS and remove .git extension
	url = strings.Replace(url, "ssh://git@", "https://", 1)
	url = strings.Replace(url, "git@", "https://", 1)
	url = strings.TrimSuffix(url, ".git")
	return url
}
