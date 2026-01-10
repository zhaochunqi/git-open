package cmd

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
)

// HostingService represents the type of Git hosting service.
type HostingService int

const (
	Unknown HostingService = iota
	GitHub
	GitLab
	Bitbucket
	// Add other services as needed
)

// getCurrentGitDirectoryFunc is a variable that can be replaced for testing
var getCurrentGitDirectoryFunc = func() (*git.Repository, error) {
	// Open the Git repository in the current working directory or any parent directory
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func getCurrentGitDirectory() (*git.Repository, error) {
	return getCurrentGitDirectoryFunc()
}

// getRemoteURLFunc is a variable that can be replaced for testing
var getRemoteURLFunc = func(repo *git.Repository) (string, error) {
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

func getRemoteURL(repo *git.Repository) (string, error) {
	return getRemoteURLFunc(repo)
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

// getBranchNameFunc is a variable that can be replaced for testing
var getBranchNameFunc = func(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err == nil {
		return head.Name().Short(), nil
	}

	ref, refErr := repo.Reference("HEAD", true)
	if refErr != nil {
		return "", fmt.Errorf("error getting HEAD: %w", err)
	}

	target := ref.Target()
	if target.IsBranch() {
		return target.Short(), nil
	}

	return "", fmt.Errorf("error getting HEAD: %w", err)
}

func getBranchName(repo *git.Repository) (string, error) {
	return getBranchNameFunc(repo)
}

// getHostingService determines the Git hosting service from the remote URL.
func getHostingService(remoteURL string) HostingService {
	if strings.Contains(remoteURL, "github.com") {
		return GitHub
	}
	if strings.Contains(remoteURL, "gitlab.com") {
		return GitLab
	}
	if strings.Contains(remoteURL, "bitbucket.org") {
		return Bitbucket
	}
	return Unknown
}

// buildBranchURL constructs the full URL for a given branch based on the hosting service.
func buildBranchURL(baseURL, branchName, remoteURL string) string {
	service := getHostingService(remoteURL)
	switch service {
	case GitHub:
		return fmt.Sprintf("%s/tree/%s", baseURL, branchName)
	case GitLab:
		return fmt.Sprintf("%s/-/tree/%s", baseURL, branchName)
	case Bitbucket:
		return fmt.Sprintf("%s/src/%s", baseURL, branchName)
	default:
		// Default to GitHub-like path for unknown services or if no specific path is needed
		return fmt.Sprintf("%s/tree/%s", baseURL, branchName)
	}
}
