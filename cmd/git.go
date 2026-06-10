package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
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
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit:          true,
		EnableDotGitCommonDir: true,
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

var scpRemoteURLPattern = regexp.MustCompile(`^(?:[^@]+@)?([^:]+):(.+)$`)

func convertToWebURL(rawURL string) string {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return ""
	}

	parsedURL, err := url.Parse(raw)
	if err == nil && parsedURL.Host != "" && parsedURL.Scheme != "" {
		// URL style: https://, http://, ssh:// or git+ssh://
		path := strings.TrimPrefix(parsedURL.Path, "/")
		if path == "" {
			return ""
		}

		switch parsedURL.Scheme {
		case "http", "https":
			host := parsedURL.Host
			return fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, strings.TrimSuffix(host, "/"), strings.TrimSuffix(path, ".git"))
		case "ssh", "git+ssh":
			return fmt.Sprintf("https://%s/%s", parsedURL.Hostname(), strings.TrimSuffix(path, ".git"))
		}

		return ""
	}

	matches := scpRemoteURLPattern.FindStringSubmatch(raw)
	if len(matches) != 3 {
		return ""
	}

	host := matches[1]
	path := strings.TrimPrefix(matches[2], "/")
	if host == "" || path == "" {
		return ""
	}

	return fmt.Sprintf("https://%s/%s", host, strings.TrimSuffix(path, ".git"))
}

// getBranchNameFunc is a variable that can be replaced for testing
var getBranchNameFunc = func(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err == nil {
		if head.Name().IsBranch() {
			return head.Name().Short(), nil
		}
		err = errors.New("detached HEAD")
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
