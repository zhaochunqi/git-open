package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
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
	if err == nil {
		return repo, nil
	}
	if !isRepositoryFormatError(err) {
		return nil, err
	}

	// go-git rejects repositories that enable config extensions it does not
	// know about (e.g. extensions.worktreeConfig), even though plain git
	// handles them fine. Since git-open only reads the remote URL and the
	// current branch, fall back to opening the repository while ignoring the
	// extensions section of the config.
	return openRepositoryToleratingExtensions(".")
}

// isRepositoryFormatError reports whether err is one of go-git's strict
// repository format/extension validation errors.
func isRepositoryFormatError(err error) bool {
	return errors.Is(err, git.ErrUnsupportedExtensionRepositoryFormatVersion) ||
		errors.Is(err, git.ErrUnknownExtension) ||
		errors.Is(err, git.ErrUnsupportedRepositoryFormatVersion)
}

// extensionTolerantStorer wraps a storage.Storer and hides the [extensions]
// section of the repository config, so go-git's extension validation passes
// for extensions it does not implement.
type extensionTolerantStorer struct {
	storage.Storer
}

func (s extensionTolerantStorer) Config() (*config.Config, error) {
	cfg, err := s.Storer.Config()
	if err != nil {
		return nil, err
	}
	if cfg != nil && cfg.Raw != nil {
		cfg.Raw.RemoveSection("extensions")
	}
	return cfg, nil
}

// openRepositoryToleratingExtensions mirrors git.PlainOpenWithOptions with
// DetectDotGit and EnableDotGitCommonDir enabled, but skips go-git's
// extension validation by hiding the [extensions] config section.
func openRepositoryToleratingExtensions(path string) (*git.Repository, error) {
	dot, wt, err := dotGitFilesystems(path)
	if err != nil {
		return nil, err
	}

	if _, err := dot.Stat(""); err != nil {
		if os.IsNotExist(err) {
			return nil, git.ErrRepositoryNotExists
		}
		return nil, err
	}

	repositoryFs := dot
	if common, err := dotGitCommonDirectory(dot); err != nil {
		return nil, err
	} else if common != nil {
		repositoryFs = dotgit.NewRepositoryFilesystem(dot, common)
	}

	s := filesystem.NewStorage(repositoryFs, cache.NewObjectLRUDefault())
	return git.Open(extensionTolerantStorer{Storer: s}, wt)
}

// dotGitFilesystems locates the .git filesystem and worktree filesystem for
// path, walking up parent directories. It handles both .git directories and
// .git files ("gitdir: <path>") used by worktrees and submodules.
func dotGitFilesystems(path string) (dot, wt billy.Filesystem, err error) {
	if path, err = filepath.Abs(path); err != nil {
		return nil, nil, err
	}

	var fs billy.Filesystem
	var fi os.FileInfo
	for {
		fs = osfs.New(path)
		fi, err = fs.Stat(git.GitDirName)
		if err == nil {
			break
		}
		if !os.IsNotExist(err) {
			return nil, nil, err
		}
		if dir := filepath.Dir(path); dir != path {
			path = dir
			continue
		}
		return nil, nil, git.ErrRepositoryNotExists
	}

	if fi.IsDir() {
		dot, err = fs.Chroot(git.GitDirName)
		return dot, fs, err
	}

	dot, err = dotGitFileToFilesystem(path, fs)
	if err != nil {
		return nil, nil, err
	}
	return dot, fs, nil
}

// dotGitFileToFilesystem resolves a .git file to the git directory it points to.
func dotGitFileToFilesystem(path string, fs billy.Filesystem) (billy.Filesystem, error) {
	f, err := fs.Open(git.GitDirName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	line := string(b)
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return nil, fmt.Errorf(".git file has no %s prefix", prefix)
	}

	gitdir := strings.TrimSpace(strings.Split(line[len(prefix):], "\n")[0])
	if filepath.IsAbs(gitdir) {
		return osfs.New(gitdir), nil
	}
	return osfs.New(fs.Join(path, gitdir)), nil
}

// dotGitCommonDirectory returns the common git directory referenced by a
// "commondir" file, or nil when the file does not exist.
func dotGitCommonDirectory(fs billy.Filesystem) (billy.Filesystem, error) {
	f, err := fs.Open("commondir")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, nil
	}

	common := strings.TrimSpace(string(b))
	var commonDir billy.Filesystem
	if filepath.IsAbs(common) {
		commonDir = osfs.New(common)
	} else {
		commonDir = osfs.New(filepath.Join(fs.Root(), common))
	}
	if _, err := commonDir.Stat(""); err != nil {
		if os.IsNotExist(err) {
			return nil, git.ErrRepositoryIncomplete
		}
		return nil, err
	}
	return commonDir, nil
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

// resolveWebURL returns the repository, its remote URL, and the converted web URL
// for the Git repository in the current working directory.
func resolveWebURL() (*git.Repository, string, string, error) {
	repo, err := getCurrentGitDirectory()
	if err != nil {
		return nil, "", "", fmt.Errorf("error getting git directory: %w", err)
	}

	remoteURL, err := getRemoteURL(repo)
	if err != nil {
		return nil, "", "", fmt.Errorf("error getting remote URL: %w", err)
	}

	webURL := convertToWebURL(remoteURL)
	if webURL == "" {
		return nil, "", "", fmt.Errorf("unsupported remote URL format: %s", remoteURL)
	}

	return repo, remoteURL, webURL, nil
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
