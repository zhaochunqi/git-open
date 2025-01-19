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
	// Validate URL format
	if !strings.Contains(url, "://") && !strings.Contains(url, "@") {
		return ""
	}

	// If the URL starts with "https://" or "http://", remove the ".git" suffix
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		url = strings.TrimSuffix(url, ".git")
	} else {
		// Otherwise, assume it's an SSH URL
		// Remove "ssh://" prefix
		url = strings.TrimPrefix(url, "ssh://")
		url = strings.Replace(url, ":", "/", 1)
		// Replace "git@" or "ssh://git@" with "https://"
		url = strings.Replace(url, "git@", "https://", 1)
		// Remove the ".git" suffix
		url = strings.TrimSuffix(url, ".git")
	}
	return url
}
