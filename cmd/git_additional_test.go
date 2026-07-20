package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

func Test_getBranchName_DetachedHEAD(t *testing.T) {
	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
	defer cleanup()

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error = %v", err)
	}

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("repo.Head() error = %v", err)
	}

	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.HEAD, head.Hash())); err != nil {
		t.Fatalf("set detached HEAD failed: %v", err)
	}

	branch, err := getBranchName(repo)
	if err == nil {
		t.Fatalf("getBranchName() expected error on detached HEAD, got branch=%q", branch)
	}
	if !strings.Contains(err.Error(), "error getting HEAD") {
		t.Fatalf("getBranchName() error = %v, want message containing 'error getting HEAD'", err)
	}
}

// Regression test: repositories that enable config extensions unknown to
// go-git (e.g. extensions.worktreeConfig) must still open successfully.
func Test_resolveWebURL_WorktreeConfigExtension(t *testing.T) {
	tmpDir, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	configPath := filepath.Join(tmpDir, ".git", "config")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open config failed: %v", err)
	}
	if _, err := f.WriteString("[extensions]\n\tworktreeConfig = true\n"); err != nil {
		f.Close()
		t.Fatalf("append extensions section failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close config failed: %v", err)
	}

	repo, remoteURL, webURL, err := resolveWebURL()
	if err != nil {
		t.Fatalf("resolveWebURL() error = %v", err)
	}
	if remoteURL != "https://github.com/zhaochunqi/git-open.git" {
		t.Errorf("resolveWebURL() remoteURL = %q, want %q", remoteURL, "https://github.com/zhaochunqi/git-open.git")
	}
	if webURL != "https://github.com/zhaochunqi/git-open" {
		t.Errorf("resolveWebURL() webURL = %q, want %q", webURL, "https://github.com/zhaochunqi/git-open")
	}

	branch, err := getBranchName(repo)
	if err != nil {
		t.Fatalf("getBranchName() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("getBranchName() = %q, want %q", branch, "main")
	}
}

// Regression test: the extension-tolerant fallback must also work from a
// linked worktree, where .git is a file pointing to a git directory that
// references a commondir.
func Test_openRepositoryToleratingExtensions_Worktree(t *testing.T) {
	worktreeDir, cleanup := testhelper.SetupTestWorktree(t, "https://github.com/zhaochunqi/git-open.git", "feature-branch")
	defer cleanup()

	// Locate the main repository's common git directory through the
	// worktree's .git file and its commondir pointer.
	b, err := os.ReadFile(filepath.Join(worktreeDir, ".git"))
	if err != nil {
		t.Fatalf("read .git file failed: %v", err)
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(string(b), "gitdir: "))
	cb, err := os.ReadFile(filepath.Join(gitdir, "commondir"))
	if err != nil {
		t.Fatalf("read commondir failed: %v", err)
	}
	commonDir := strings.TrimSpace(string(cb))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(gitdir, commonDir)
	}

	// Enable an extension unknown to go-git in the common config, so the
	// strict PlainOpenWithOptions path fails and the fallback is exercised.
	configPath := filepath.Join(commonDir, "config")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open config failed: %v", err)
	}
	if _, err := f.WriteString("[extensions]\n\tworktreeConfig = true\n"); err != nil {
		f.Close()
		t.Fatalf("append extensions section failed: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close config failed: %v", err)
	}

	repo, err := getCurrentGitDirectory()
	if err != nil {
		t.Fatalf("getCurrentGitDirectory() error in worktree with extensions = %v", err)
	}

	branch, err := getBranchName(repo)
	if err != nil {
		t.Fatalf("getBranchName() error = %v", err)
	}
	if branch != "feature-branch" {
		t.Errorf("getBranchName() = %q, want %q", branch, "feature-branch")
	}

	remoteURL, err := getRemoteURL(repo)
	if err != nil {
		t.Fatalf("getRemoteURL() error = %v", err)
	}
	if remoteURL != "https://github.com/zhaochunqi/git-open.git" {
		t.Errorf("getRemoteURL() = %q, want %q", remoteURL, "https://github.com/zhaochunqi/git-open.git")
	}
}

func Test_dotGitFilesystems_NotFound(t *testing.T) {
	// A plain temporary directory has no .git anywhere up to the filesystem
	// root, so the lookup must report ErrRepositoryNotExists.
	tmpDir := t.TempDir()

	_, _, err := dotGitFilesystems(tmpDir)
	if !errors.Is(err, git.ErrRepositoryNotExists) {
		t.Errorf("dotGitFilesystems() error = %v, want %v", err, git.ErrRepositoryNotExists)
	}
}

func Test_dotGitFileToFilesystem(t *testing.T) {
	absGitdir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantRoot string
		wantErr  bool
	}{
		{
			name:     "relative gitdir",
			content:  "gitdir: .git-real\n",
			wantRoot: ".git-real",
		},
		{
			name:     "absolute gitdir",
			content:  "gitdir: " + absGitdir + "\n",
			wantRoot: absGitdir,
		},
		{
			name:    "missing gitdir prefix",
			content: "not a gitdir pointer\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, ".git"), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			fs, err := dotGitFileToFilesystem(tmpDir, osfs.New(tmpDir))
			if (err != nil) != tt.wantErr {
				t.Fatalf("dotGitFileToFilesystem() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			want := tt.wantRoot
			if !filepath.IsAbs(want) {
				want = filepath.Join(tmpDir, want)
			}
			// osfs.Root resolves symlinks (e.g. /var -> /private/var on macOS).
			if resolved, err := filepath.EvalSymlinks(want); err == nil {
				want = resolved
			}
			if fs.Root() != want {
				t.Errorf("dotGitFileToFilesystem() root = %q, want %q", fs.Root(), want)
			}
		})
	}
}

func Test_dotGitFileToFilesystem_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := dotGitFileToFilesystem(tmpDir, osfs.New(tmpDir))
	if err == nil {
		t.Error("dotGitFileToFilesystem() expected error for missing .git file, got nil")
	}
}

func Test_dotGitCommonDirectory(t *testing.T) {
	t.Run("no commondir file", func(t *testing.T) {
		fs := osfs.New(t.TempDir())
		common, err := dotGitCommonDirectory(fs)
		if err != nil {
			t.Fatalf("dotGitCommonDirectory() error = %v", err)
		}
		if common != nil {
			t.Errorf("dotGitCommonDirectory() = %v, want nil", common)
		}
	})

	t.Run("empty commondir file", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "commondir"), nil, 0644); err != nil {
			t.Fatal(err)
		}
		common, err := dotGitCommonDirectory(osfs.New(tmpDir))
		if err != nil {
			t.Fatalf("dotGitCommonDirectory() error = %v", err)
		}
		if common != nil {
			t.Errorf("dotGitCommonDirectory() = %v, want nil", common)
		}
	})

	t.Run("relative commondir", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitdir := filepath.Join(tmpDir, "gitdir")
		commonDir := filepath.Join(tmpDir, "common")
		for _, dir := range []string{gitdir, commonDir} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(gitdir, "commondir"), []byte("../common\n"), 0644); err != nil {
			t.Fatal(err)
		}
		common, err := dotGitCommonDirectory(osfs.New(gitdir))
		if err != nil {
			t.Fatalf("dotGitCommonDirectory() error = %v", err)
		}
		// osfs.Root resolves symlinks (e.g. /var -> /private/var on macOS).
		wantDir, err := filepath.EvalSymlinks(commonDir)
		if err != nil {
			t.Fatal(err)
		}
		if common == nil || common.Root() != wantDir {
			t.Errorf("dotGitCommonDirectory() root = %v, want %q", common, wantDir)
		}
	})

	t.Run("absolute commondir", func(t *testing.T) {
		tmpDir := t.TempDir()
		commonDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "commondir"), []byte(commonDir+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		common, err := dotGitCommonDirectory(osfs.New(tmpDir))
		if err != nil {
			t.Fatalf("dotGitCommonDirectory() error = %v", err)
		}
		// osfs.Root resolves symlinks (e.g. /var -> /private/var on macOS).
		wantDir, err := filepath.EvalSymlinks(commonDir)
		if err != nil {
			t.Fatal(err)
		}
		if common == nil || common.Root() != wantDir {
			t.Errorf("dotGitCommonDirectory() root = %v, want %q", common, wantDir)
		}
	})

	t.Run("commondir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "commondir"), []byte("missing\n"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := dotGitCommonDirectory(osfs.New(tmpDir))
		if !errors.Is(err, git.ErrRepositoryIncomplete) {
			t.Errorf("dotGitCommonDirectory() error = %v, want %v", err, git.ErrRepositoryIncomplete)
		}
	})
}
