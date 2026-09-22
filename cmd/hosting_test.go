package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/zhaochunqi/git-open/internal/testhelper"
)

// withHostBranchPaths swaps the global override map for the duration of a
// test and restores it afterwards, keeping tests independent of each other.
func withHostBranchPaths(t *testing.T, paths map[string]string) {
	t.Helper()
	original := HostBranchPaths
	HostBranchPaths = paths
	t.Cleanup(func() { HostBranchPaths = original })
}

// withDefaultBranchPath swaps the global default branch path for a test.
func withDefaultBranchPath(t *testing.T, path string) {
	t.Helper()
	original := DefaultBranchPath
	DefaultBranchPath = path
	t.Cleanup(func() { DefaultBranchPath = original })
}

func TestHostingServiceString(t *testing.T) {
	tests := []struct {
		service HostingService
		want    string
	}{
		{Unknown, "unknown"},
		{GitHub, "github"},
		{GitLab, "gitlab"},
		{Bitbucket, "bitbucket"},
		{Gitea, "gitea"},
		{SourceHut, "sourcehut"},
		{AzureDevOps, "azure-devops"},
		{HostingService(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.service.String(); got != tt.want {
				t.Errorf("HostingService(%d).String() = %q, want %q", tt.service, got, tt.want)
			}
		})
	}
}

func Test_hostFromRemoteURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"https", "https://github.com/user/repo.git", "github.com"},
		{"http", "http://gitlab.com/user/repo", "gitlab.com"},
		{"ssh scheme", "ssh://git@github.com/user/repo.git", "github.com"},
		{"ssh scheme with port", "ssh://git@github.com:22/user/repo.git", "github.com"},
		{"git+ssh scheme", "git+ssh://git@gitlab.com/user/repo.git", "gitlab.com"},
		{"scp with user", "git@github.com:user/repo.git", "github.com"},
		{"scp without user", "github.com:user/repo.git", "github.com"},
		{"scp with port-less custom host", "git@gitlab.example.com:group/sub/repo.git", "gitlab.example.com"},
		{"surrounding whitespace", "  https://github.com/user/repo.git  ", "github.com"},
		{"hostname lowercased", "https://GitHub.COM/User/Repo.git", "github.com"},
		{"no path", "https://example.com", "example.com"},
		{"empty", "", ""},
		{"blank", "   ", ""},
		{"no host", "invalid-remote", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hostFromRemoteURL(tt.url); got != tt.want {
				t.Errorf("hostFromRemoteURL(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func Test_hostHasLabel(t *testing.T) {
	tests := []struct {
		hostname string
		label    string
		want     bool
	}{
		{"gitlab.example.com", "gitlab", true},
		{"code.gitlab.example.com", "gitlab", true},
		{"example.com", "gitlab", false},
		{"notgitlab.com", "gitlab", false},
		{"mygitlab.com", "gitlab", false},
		{"gitlab.com", "gitlab", true},
		{"", "gitlab", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostname+"/"+tt.label, func(t *testing.T) {
			if got := hostHasLabel(tt.hostname, tt.label); got != tt.want {
				t.Errorf("hostHasLabel(%q, %q) = %v, want %v", tt.hostname, tt.label, got, tt.want)
			}
		})
	}
}

func Test_getHostingService(t *testing.T) {
	tests := []struct {
		name      string
		remoteURL string
		want      HostingService
	}{
		{"github https", "https://github.com/user/repo.git", GitHub},
		{"github scp", "git@github.com:user/repo.git", GitHub},
		{"gitlab", "https://gitlab.com/user/repo.git", GitLab},
		{"bitbucket", "https://bitbucket.org/user/repo.git", Bitbucket},
		{"codeberg", "https://codeberg.org/user/repo.git", Gitea},
		{"gitea", "https://gitea.com/user/repo.git", Gitea},
		{"sourcehut", "https://git.sr.ht/~user/repo", SourceHut},
		{"azure devops", "https://dev.azure.com/org/project/_git/repo", AzureDevOps},
		{"visual studio suffix", "https://org.visualstudio.com/project/_git/repo", AzureDevOps},
		{"self-hosted gitlab", "https://gitlab.example.com/group/repo.git", GitLab},
		{"nested gitlab label", "https://code.gitlab.example.com/group/repo.git", GitLab},
		{"self-hosted gitea", "git@gitea.example.com:user/repo.git", Gitea},
		{"self-hosted forgejo", "https://forgejo.example.com/user/repo.git", Gitea},
		// Regression: the old implementation used strings.Contains on the whole
		// remote URL, so a path mentioning github.com was misclassified.
		{"path contains github.com", "https://git.example.com/mirror/github.com.git", Unknown},
		// Regression: a hostname merely containing "gitlab" must not match.
		{"hostname contains gitlab", "https://notgitlab.com/user/repo.git", Unknown},
		{"github-like hostname", "https://github.company.com/user/repo.git", Unknown},
		{"arbitrary enterprise host", "https://git.mycorp.com/user/repo.git", Unknown},
		{"empty", "", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHostingService(tt.remoteURL); got != tt.want {
				t.Errorf("getHostingService(%q) = %v, want %v", tt.remoteURL, got, tt.want)
			}
		})
	}
}

func Test_branchPathForHost(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		override    map[string]string
		defaultPath string
		want        string
	}{
		{"github", "github.com", nil, "", "/tree/{branch}"},
		{"gitlab", "gitlab.com", nil, "", "/-/tree/{branch}"},
		{"bitbucket", "bitbucket.org", nil, "", "/src/{branch}"},
		{"codeberg", "codeberg.org", nil, "", "/src/branch/{branch}"},
		{"sourcehut", "git.sr.ht", nil, "", "/tree/{branch}"},
		{"azure devops", "dev.azure.com", nil, "", "?version=GB{branch}"},
		{"self-hosted gitlab", "gitlab.example.com", nil, "", "/-/tree/{branch}"},
		{"unknown host", "example.com", nil, "", ""},
		{"empty host", "", nil, "/tree/{branch}", ""},
		{"override wins over built-in", "github.com", map[string]string{"github.com": "/custom/{branch}"}, "", "/custom/{branch}"},
		{"override adds unknown host", "git.example.com", map[string]string{"git.example.com": "/-/tree/{branch}"}, "", "/-/tree/{branch}"},
		{"empty override forces root", "github.com", map[string]string{"github.com": ""}, "/tree/{branch}", ""},
		{"default style for unknown host", "git.example.com", nil, "/src/branch/{branch}", "/src/branch/{branch}"},
		{"default does not beat built-in", "github.com", nil, "/src/branch/{branch}", "/tree/{branch}"},
		{"default does not beat override", "git.example.com", map[string]string{"git.example.com": "/-/tree/{branch}"}, "/src/branch/{branch}", "/-/tree/{branch}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withHostBranchPaths(t, tt.override)
			withDefaultBranchPath(t, tt.defaultPath)
			if got := branchPathForHost(tt.hostname); got != tt.want {
				t.Errorf("branchPathForHost(%q) = %q, want %q", tt.hostname, got, tt.want)
			}
		})
	}
}

func Test_buildBranchURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		branch      string
		remoteURL   string
		override    map[string]string
		defaultPath string
		want        string
	}{
		{
			name:      "github",
			baseURL:   "https://github.com/user/repo",
			branch:    "feature",
			remoteURL: "https://github.com/user/repo.git",
			want:      "https://github.com/user/repo/tree/feature",
		},
		{
			name:      "gitlab",
			baseURL:   "https://gitlab.com/user/repo",
			branch:    "feature",
			remoteURL: "https://gitlab.com/user/repo.git",
			want:      "https://gitlab.com/user/repo/-/tree/feature",
		},
		{
			name:      "bitbucket",
			baseURL:   "https://bitbucket.org/user/repo",
			branch:    "feature",
			remoteURL: "https://bitbucket.org/user/repo.git",
			want:      "https://bitbucket.org/user/repo/src/feature",
		},
		{
			name:      "codeberg",
			baseURL:   "https://codeberg.org/user/repo",
			branch:    "feature",
			remoteURL: "https://codeberg.org/user/repo.git",
			want:      "https://codeberg.org/user/repo/src/branch/feature",
		},
		{
			name:      "sourcehut",
			baseURL:   "https://git.sr.ht/~user/repo",
			branch:    "feature",
			remoteURL: "https://git.sr.ht/~user/repo",
			want:      "https://git.sr.ht/~user/repo/tree/feature",
		},
		{
			name:      "azure devops",
			baseURL:   "https://dev.azure.com/org/project/_git/repo",
			branch:    "feature",
			remoteURL: "https://dev.azure.com/org/project/_git/repo",
			want:      "https://dev.azure.com/org/project/_git/repo?version=GBfeature",
		},
		{
			name:      "self-hosted gitlab via scp",
			baseURL:   "https://gitlab.example.com/group/repo",
			branch:    "feature",
			remoteURL: "git@gitlab.example.com:group/repo.git",
			want:      "https://gitlab.example.com/group/repo/-/tree/feature",
		},
		{
			name:      "branch name with slash",
			baseURL:   "https://github.com/user/repo",
			branch:    "feature/nested",
			remoteURL: "https://github.com/user/repo.git",
			want:      "https://github.com/user/repo/tree/feature/nested",
		},
		{
			name:      "unknown host falls back to repository root",
			baseURL:   "https://git.mycorp.com/user/repo",
			branch:    "feature",
			remoteURL: "https://git.mycorp.com/user/repo.git",
			want:      "https://git.mycorp.com/user/repo",
		},
		{
			name:      "unknown host with scp remote falls back to root",
			baseURL:   "https://git.mycorp.com/user/repo",
			branch:    "feature",
			remoteURL: "git@git.mycorp.com:user/repo.git",
			want:      "https://git.mycorp.com/user/repo",
		},
		{
			name:      "user override for unknown host",
			baseURL:   "https://git.mycorp.com/user/repo",
			branch:    "feature",
			remoteURL: "https://git.mycorp.com/user/repo.git",
			override:  map[string]string{"git.mycorp.com": "/-/tree/{branch}"},
			want:      "https://git.mycorp.com/user/repo/-/tree/feature",
		},
		{
			name:      "user override wins over built-in",
			baseURL:   "https://github.com/user/repo",
			branch:    "feature",
			remoteURL: "https://github.com/user/repo.git",
			override:  map[string]string{"github.com": "/custom/{branch}"},
			want:      "https://github.com/user/repo/custom/feature",
		},
		{
			name:      "empty override disables branch link",
			baseURL:   "https://github.com/user/repo",
			branch:    "feature",
			remoteURL: "https://github.com/user/repo.git",
			override:  map[string]string{"github.com": ""},
			want:      "https://github.com/user/repo",
		},
		{
			name:        "default style applies to unknown host",
			baseURL:     "https://git.mycorp.com/user/repo",
			branch:      "feature",
			remoteURL:   "https://git.mycorp.com/user/repo.git",
			defaultPath: "/src/branch/{branch}",
			want:        "https://git.mycorp.com/user/repo/src/branch/feature",
		},
		{
			name:        "default style does not override a built-in rule",
			baseURL:     "https://github.com/user/repo",
			branch:      "feature",
			remoteURL:   "https://github.com/user/repo.git",
			defaultPath: "/src/branch/{branch}",
			want:        "https://github.com/user/repo/tree/feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withHostBranchPaths(t, tt.override)
			withDefaultBranchPath(t, tt.defaultPath)
			if got := buildBranchURL(tt.baseURL, tt.branch, tt.remoteURL); got != tt.want {
				t.Errorf("buildBranchURL(%q, %q, %q) = %q, want %q", tt.baseURL, tt.branch, tt.remoteURL, got, tt.want)
			}
		})
	}
}

func Test_parseHostBranchPaths(t *testing.T) {
	tests := []struct {
		name    string
		raw     map[string]any
		want    map[string]string
		wantErr string
	}{
		{
			name: "plain string values",
			raw: map[string]any{
				"gitlab.example.com": "/-/tree/{branch}",
				"gitea.example.com":  "/src/branch/{branch}",
			},
			want: map[string]string{
				"gitlab.example.com": "/-/tree/{branch}",
				"gitea.example.com":  "/src/branch/{branch}",
			},
		},
		{
			name: "nested branch objects",
			raw: map[string]any{
				"gitlab.example.com": map[string]any{"branch": "/-/tree/{branch}"},
				"gitea.example.com":  map[any]any{"branch": "/src/branch/{branch}"},
			},
			want: map[string]string{
				"gitlab.example.com": "/-/tree/{branch}",
				"gitea.example.com":  "/src/branch/{branch}",
			},
		},
		{
			name: "hosts are lowercased and trimmed",
			raw: map[string]any{
				"  GitLab.Example.COM  ": " /-/tree/{branch} ",
			},
			want: map[string]string{
				"gitlab.example.com": "/-/tree/{branch}",
			},
		},
		{
			name: "style names are resolved to templates",
			raw: map[string]any{
				"git.example.com":      "gitea",
				"forgejo.example.com":  "forgejo",
				"codeberg.example.com": "codeberg",
				"lab.example.com":      "gitlab",
				"hub.example.com":      "github",
			},
			want: map[string]string{
				"git.example.com":      "/src/branch/{branch}",
				"forgejo.example.com":  "/src/branch/{branch}",
				"codeberg.example.com": "/src/branch/{branch}",
				"lab.example.com":      "/-/tree/{branch}",
				"hub.example.com":      "/tree/{branch}",
			},
		},
		{
			name: "style names are case-insensitive",
			raw: map[string]any{
				"git.example.com": "Gitea",
			},
			want: map[string]string{
				"git.example.com": "/src/branch/{branch}",
			},
		},
		{
			name: "nested style objects",
			raw: map[string]any{
				"gitlab.example.com": map[string]any{"style": "gitlab"},
				"gitea.example.com":  map[any]any{"style": "gitea"},
			},
			want: map[string]string{
				"gitlab.example.com": "/-/tree/{branch}",
				"gitea.example.com":  "/src/branch/{branch}",
			},
		},
		{
			name: "branch key wins over style key",
			raw: map[string]any{
				"git.example.com": map[string]any{"style": "gitea", "branch": "/custom/{branch}"},
			},
			want: map[string]string{
				"git.example.com": "/custom/{branch}",
			},
		},
		{
			name: "none and empty force the repository root",
			raw: map[string]any{
				"a.example.com": "none",
				"b.example.com": "",
				"c.example.com": map[string]any{"style": "none"},
			},
			want: map[string]string{
				"a.example.com": "",
				"b.example.com": "",
				"c.example.com": "",
			},
		},
		{
			name: "empty config",
			raw:  map[string]any{},
			want: map[string]string{},
		},
		{
			name:    "blank hostname key",
			raw:     map[string]any{"": "/tree/{branch}"},
			wantErr: "empty hostname key",
		},
		{
			name:    "whitespace-only hostname key",
			raw:     map[string]any{"   ": "/tree/{branch}"},
			wantErr: "empty hostname key",
		},
		{
			name:    "non-string value",
			raw:     map[string]any{"a.example.com": 42},
			wantErr: `hosts["a.example.com"]: expected`,
		},
		{
			name:    "nil value",
			raw:     map[string]any{"a.example.com": nil},
			wantErr: `hosts["a.example.com"]: expected`,
		},
		{
			name:    "nested object without style or branch",
			raw:     map[string]any{"a.example.com": map[string]any{"other": "/x"}},
			wantErr: `hosts["a.example.com"]: expected`,
		},
		{
			name:    "nested branch is not a string",
			raw:     map[string]any{"a.example.com": map[string]any{"branch": 42}},
			wantErr: `hosts["a.example.com"]: expected`,
		},
		{
			name:    "unknown style in a plain value",
			raw:     map[string]any{"a.example.com": "bogus-style"},
			wantErr: `hosts["a.example.com"]: unknown style "bogus-style"`,
		},
		{
			name:    "unknown style in a nested object",
			raw:     map[string]any{"a.example.com": map[string]any{"style": "bogus-style"}},
			wantErr: `hosts["a.example.com"]: unknown style "bogus-style"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHostBranchPaths(tt.raw)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("parseHostBranchPaths() error = %v, want error containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHostBranchPaths() unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("parseHostBranchPaths() = %v, want %v", got, tt.want)
			}
			for host, want := range tt.want {
				if got[host] != want {
					t.Errorf("parseHostBranchPaths()[%q] = %q, want %q", host, got[host], want)
				}
			}
		})
	}
}

// runRootForRepo runs the root command's RunE against a repository with the
// given remote/branch and returns the URL that would have been opened.
func runRootForRepo(t *testing.T, remoteURL, branch string) string {
	t.Helper()

	_, cleanup := testhelper.SetupTestRepo(t, remoteURL, branch)
	t.Cleanup(cleanup)

	original := OpenURLInBrowser
	t.Cleanup(func() { OpenURLInBrowser = original })

	var opened string
	OpenURLInBrowser = func(url string) error {
		opened = url
		return nil
	}

	cmd := &cobra.Command{}
	cmd.SetOut(new(bytes.Buffer))
	cmd.Flags().Bool("plain", false, "")

	if err := rootCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("rootCmd.RunE() error = %v", err)
	}
	return opened
}

// Test_rootCmd_UnknownHostFallsBackToRoot is the end-to-end guarantee behind
// "any repository": a host with no built-in rule opens the repository root
// rather than a guessed branch URL that would 404.
func Test_rootCmd_UnknownHostFallsBackToRoot(t *testing.T) {
	withHostBranchPaths(t, nil)
	withDefaultBranchPath(t, "")

	got := runRootForRepo(t, "https://git.mycorp.com/user/repo.git", "feature-branch")
	if want := "https://git.mycorp.com/user/repo"; got != want {
		t.Errorf("opened URL = %q, want %q", got, want)
	}
}

// Test_rootCmd_HostOverride ensures the hosts config reaches the root command.
func Test_rootCmd_HostOverride(t *testing.T) {
	withHostBranchPaths(t, map[string]string{"git.mycorp.com": "/-/tree/{branch}"})

	got := runRootForRepo(t, "https://git.mycorp.com/user/repo.git", "feature-branch")
	if want := "https://git.mycorp.com/user/repo/-/tree/feature-branch"; got != want {
		t.Errorf("opened URL = %q, want %q", got, want)
	}
}

func Test_stylePath(t *testing.T) {
	tests := []struct {
		name string
		want string
		ok   bool
	}{
		{"github", "/tree/{branch}", true},
		{"gitlab", "/-/tree/{branch}", true},
		{"bitbucket", "/src/{branch}", true},
		{"gitea", "/src/branch/{branch}", true},
		{"forgejo", "/src/branch/{branch}", true},
		{"codeberg", "/src/branch/{branch}", true},
		{"sourcehut", "/tree/{branch}", true},
		{"azure-devops", "?version=GB{branch}", true},
		{"none", "", true},
		{"  GITLAB  ", "/-/tree/{branch}", true},
		{"bogus", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := stylePath(tt.name)
			if got != tt.want || ok != tt.ok {
				t.Errorf("stylePath(%q) = (%q, %v), want (%q, %v)", tt.name, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func Test_resolveBranchSpec(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    string
		wantErr string
	}{
		{"style name", "gitea", "/src/branch/{branch}", ""},
		{"raw template", "/custom/{branch}/x", "/custom/{branch}/x", ""},
		{"whitespace around style", "  gitlab ", "/-/tree/{branch}", ""},
		{"empty is root", "", "", ""},
		{"whitespace only is root", "   ", "", ""},
		{"none is root", "none", "", ""},
		{"unknown style", "bogus", "", `unknown style "bogus"`},
		{"unknown style lists known ones", "bogus", "", "known styles:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBranchSpec(tt.spec)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("resolveBranchSpec(%q) unexpected error: %v", tt.spec, err)
				}
				if got != tt.want {
					t.Errorf("resolveBranchSpec(%q) = %q, want %q", tt.spec, got, tt.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("resolveBranchSpec(%q) error = %v, want error containing %q", tt.spec, err, tt.wantErr)
			}
		})
	}
}

func Test_parseDefaultBranchPath(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    string
		wantErr string
	}{
		{"unset", "", "", ""},
		{"style name", "gitea", "/src/branch/{branch}", ""},
		{"case-insensitive style", "GitLab", "/-/tree/{branch}", ""},
		{"none", "none", "", ""},
		{"raw template", "/x/{branch}", "/x/{branch}", ""},
		{"unknown style", "bogus", "", `default_style: unknown style "bogus"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDefaultBranchPath(tt.spec)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("parseDefaultBranchPath(%q) unexpected error: %v", tt.spec, err)
				}
				if got != tt.want {
					t.Errorf("parseDefaultBranchPath(%q) = %q, want %q", tt.spec, got, tt.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("parseDefaultBranchPath(%q) error = %v, want error containing %q", tt.spec, err, tt.wantErr)
			}
		})
	}
}

// Test_initConfigHostOverrides checks that the "hosts" section is parsed from
// the config file into HostBranchPaths.
func Test_initConfigHostOverrides(t *testing.T) {
	originalBrowserCommand := BrowserCommand
	originalHostBranchPaths := HostBranchPaths
	originalDefaultBranchPath := DefaultBranchPath
	t.Cleanup(func() {
		BrowserCommand = originalBrowserCommand
		HostBranchPaths = originalHostBranchPaths
		DefaultBranchPath = originalDefaultBranchPath
	})
	t.Setenv("BROWSER", "")

	_, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), `browser: firefox
default_style: gitea
hosts:
  gitlab.internal.example:
    branch: "/-/tree/{branch}"
  code.internal.example: "/src/branch/{branch}"
  gh.internal.example: github
  forge.internal.example:
    style: forgejo
`)

	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() error = %v", err)
	}

	if BrowserCommand != "firefox" {
		t.Errorf("BrowserCommand = %q, want %q", BrowserCommand, "firefox")
	}
	want := map[string]string{
		"gitlab.internal.example": "/-/tree/{branch}",
		"code.internal.example":   "/src/branch/{branch}",
		"gh.internal.example":     "/tree/{branch}",
		"forge.internal.example":  "/src/branch/{branch}",
	}
	if len(HostBranchPaths) != len(want) {
		t.Fatalf("HostBranchPaths = %v, want %v", HostBranchPaths, want)
	}
	for host, path := range want {
		if HostBranchPaths[host] != path {
			t.Errorf("HostBranchPaths[%q] = %q, want %q", host, HostBranchPaths[host], path)
		}
	}
	if DefaultBranchPath != "/src/branch/{branch}" {
		t.Errorf("DefaultBranchPath = %q, want %q", DefaultBranchPath, "/src/branch/{branch}")
	}
}

// Test_initConfigDefaultStyleUnset checks that a missing default_style leaves
// unknown hosts opening the repository root.
func Test_initConfigDefaultStyleUnset(t *testing.T) {
	originalDefaultBranchPath := DefaultBranchPath
	t.Cleanup(func() { DefaultBranchPath = originalDefaultBranchPath })
	t.Setenv("BROWSER", "")

	isolateConfig(t)
	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() error = %v", err)
	}

	if DefaultBranchPath != "" {
		t.Errorf("DefaultBranchPath = %q, want empty", DefaultBranchPath)
	}
}

// Test_initConfigInvalidHostStyle checks that a typo in a hosts value fails
// instead of being silently ignored.
func Test_initConfigInvalidHostStyle(t *testing.T) {
	originalHostBranchPaths := HostBranchPaths
	originalDefaultBranchPath := DefaultBranchPath
	t.Cleanup(func() {
		HostBranchPaths = originalHostBranchPaths
		DefaultBranchPath = originalDefaultBranchPath
	})
	t.Setenv("BROWSER", "")

	_, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), "hosts:\n  git.example.com: gitea2\n")

	err := initConfig()
	if err == nil {
		t.Fatal("initConfig() expected error for an unknown style, got nil")
	}
	if !strings.Contains(err.Error(), `hosts["git.example.com"]: unknown style "gitea2"`) {
		t.Errorf("initConfig() error = %v, want unknown style for hosts[git.example.com]", err)
	}
}

// Test_initConfigInvalidDefaultStyle checks that a typo in default_style is
// reported with the key name.
func Test_initConfigInvalidDefaultStyle(t *testing.T) {
	originalDefaultBranchPath := DefaultBranchPath
	t.Cleanup(func() { DefaultBranchPath = originalDefaultBranchPath })
	t.Setenv("BROWSER", "")

	_, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), "default_style: gitea2\n")

	err := initConfig()
	if err == nil {
		t.Fatal("initConfig() expected error for an unknown default_style, got nil")
	}
	if !strings.Contains(err.Error(), `default_style: unknown style "gitea2"`) {
		t.Errorf("initConfig() error = %v, want default_style unknown style", err)
	}
}

// Test_PersistentPreRunE_InvalidConfig checks that a broken config is surfaced
// by the root command's PersistentPreRunE.
func Test_PersistentPreRunE_InvalidConfig(t *testing.T) {
	originalHostBranchPaths := HostBranchPaths
	originalDefaultBranchPath := DefaultBranchPath
	t.Cleanup(func() {
		HostBranchPaths = originalHostBranchPaths
		DefaultBranchPath = originalDefaultBranchPath
	})
	t.Setenv("BROWSER", "")

	_, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), "hosts:\n  git.example.com: nope\n")

	cmd := &cobra.Command{}
	cmd.Flags().Bool("version", false, "")

	err := rootCmd.PersistentPreRunE(cmd, nil)
	if err == nil {
		t.Fatal("PersistentPreRunE() expected config error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown style") {
		t.Errorf("PersistentPreRunE() error = %v, want unknown style", err)
	}
}

// Test_PersistentPreRunE_VersionAndHelpSkipValidation checks that --version and
// the help command keep working even when the config file is broken.
func Test_PersistentPreRunE_VersionAndHelpSkipValidation(t *testing.T) {
	t.Setenv("BROWSER", "")

	_, xdg := isolateConfig(t)
	writeConfigFile(t, xdgConfigPath(xdg), "hosts:\n  git.example.com: nope\n")

	versionCmd := &cobra.Command{}
	versionCmd.Flags().Bool("version", false, "")
	if err := versionCmd.Flags().Set("version", "true"); err != nil {
		t.Fatal(err)
	}
	if err := rootCmd.PersistentPreRunE(versionCmd, nil); err != nil {
		t.Errorf("PersistentPreRunE(--version) = %v, want nil", err)
	}

	helpCmd := &cobra.Command{Use: helpCommandName}
	if err := rootCmd.PersistentPreRunE(helpCmd, nil); err != nil {
		t.Errorf("PersistentPreRunE(help) = %v, want nil", err)
	}
}
