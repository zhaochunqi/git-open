package cmd

import (
	"net/url"
	"regexp"
	"strings"
)

// HostingService identifies the Git hosting platform behind a remote URL. It
// is used for classification and, through the host rules below, to choose the
// URL shape that links to a branch.
type HostingService int

const (
	Unknown HostingService = iota
	GitHub
	GitLab
	Bitbucket
	Gitea
	SourceHut
	AzureDevOps
)

// String returns a stable, human-readable name for the service.
func (s HostingService) String() string {
	switch s {
	case GitHub:
		return "github"
	case GitLab:
		return "gitlab"
	case Bitbucket:
		return "bitbucket"
	case Gitea:
		return "gitea"
	case SourceHut:
		return "sourcehut"
	case AzureDevOps:
		return "azure-devops"
	default:
		return "unknown"
	}
}

// Branch path templates. They are shared by the built-in host rules and by the
// named styles users can reference from the config file, so the two stay in
// sync. Each template is appended to the repository root URL and may contain
// the placeholder {branch}.
const (
	branchPathGitHub      = "/tree/{branch}"
	branchPathGitLab      = "/-/tree/{branch}"
	branchPathBitbucket   = "/src/{branch}"
	branchPathGitea       = "/src/branch/{branch}"
	branchPathSourceHut   = "/tree/{branch}"
	branchPathAzureDevOps = "?version=GB{branch}"
	branchPathRoot        = ""
)

// branchStyles maps a style name to the branch path template used by that
// family of hosts. A config value that does not contain {branch} is treated as
// a style name, so "gitea" is shorthand for "/src/branch/{branch}" and "none"
// (or an empty string) forces the repository root.
var branchStyles = map[string]string{
	"github":       branchPathGitHub,
	"gitlab":       branchPathGitLab,
	"bitbucket":    branchPathBitbucket,
	"gitea":        branchPathGitea,
	"forgejo":      branchPathGitea,
	"codeberg":     branchPathGitea,
	"sourcehut":    branchPathSourceHut,
	"azure-devops": branchPathAzureDevOps,
	"none":         branchPathRoot,
}

// hostRule describes how to link to a branch on one family of hosts.
type hostRule struct {
	service HostingService
	// host matches the hostname exactly or as a dot-delimited suffix, so
	// "github.com" matches github.com and enterprise.github.com.
	host string
	// keyword matches a whole label of the hostname: "gitlab" matches
	// gitlab.example.com and code.gitlab.example.com. It is a weaker signal
	// than host, used to recognise self-hosted instances whose domain is not
	// known in advance.
	keyword string
	// branchPath is a URL suffix containing {branch}; empty means the host has
	// no predictable branch URL, so only the repository root can be opened.
	branchPath string
}

// builtinHostRules is the built-in knowledge about hosted Git services.
//
// Matching runs in three tiers: exact host, then host suffix, then keyword.
// Rules are grouped by tier so that, for example, an exact match on
// gitlab.com always wins over the "gitlab" keyword.
var builtinHostRules = []hostRule{
	// Tier 1: exact host matches (public SaaS domains).
	{GitHub, "github.com", "", branchPathGitHub},
	{GitLab, "gitlab.com", "", branchPathGitLab},
	{Bitbucket, "bitbucket.org", "", branchPathBitbucket},
	{Gitea, "codeberg.org", "", branchPathGitea},
	{Gitea, "gitea.com", "", branchPathGitea},
	{SourceHut, "git.sr.ht", "", branchPathSourceHut},
	{AzureDevOps, "dev.azure.com", "", branchPathAzureDevOps},

	// Tier 2: host suffix matches (per-tenant domains and deploy variants).
	{AzureDevOps, "visualstudio.com", "", branchPathAzureDevOps},

	// Tier 3: label keyword matches (self-hosted instances).
	{GitLab, "", "gitlab", branchPathGitLab},
	{Gitea, "", "gitea", branchPathGitea},
	{Gitea, "", "forgejo", branchPathGitea},
}

// HostBranchPaths holds the user configured host -> branch path overrides,
// loaded from the "hosts" section of the config file. Keys are lowercase
// hostnames and values are already resolved path templates. A host mapped to
// the empty string forces the repository root.
var HostBranchPaths = map[string]string{}

// DefaultBranchPath is the branch path applied to hosts with no built-in rule
// and no override, resolved from the "default_style" config key. Empty means
// unknown hosts open the repository root.
var DefaultBranchPath string

// scpRemoteURLPattern matches scp-style remotes such as git@host:owner/repo.git.
var scpRemoteURLPattern = regexp.MustCompile(`^(?:[^@]+@)?([^:]+):(.+)$`)

// hostFromRemoteURL extracts the lowercase hostname from a remote URL. It
// accepts both URL style (https://, ssh://, git+ssh://) and scp style
// (git@host:path) remotes, and returns "" when the host cannot be determined.
func hostFromRemoteURL(rawURL string) string {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return ""
	}

	if parsedURL, err := url.Parse(raw); err == nil && parsedURL.Host != "" && parsedURL.Scheme != "" {
		return strings.ToLower(parsedURL.Hostname())
	}

	matches := scpRemoteURLPattern.FindStringSubmatch(raw)
	if len(matches) != 3 {
		return ""
	}

	return strings.ToLower(matches[1])
}

// resolveHostRule returns the built-in rule that applies to hostname. The
// three matching tiers are tried in order: exact host, host suffix, then
// hostname label keyword.
func resolveHostRule(hostname string) (hostRule, bool) {
	if hostname == "" {
		return hostRule{}, false
	}

	for _, rule := range builtinHostRules {
		if rule.host != "" && rule.host == hostname {
			return rule, true
		}
	}
	for _, rule := range builtinHostRules {
		if rule.host != "" && strings.HasSuffix(hostname, "."+rule.host) {
			return rule, true
		}
	}
	for _, rule := range builtinHostRules {
		if rule.keyword != "" && hostHasLabel(hostname, rule.keyword) {
			return rule, true
		}
	}

	return hostRule{}, false
}

// hostHasLabel reports whether any dot-delimited label of hostname equals
// label. This avoids matching substrings such as "notgitlab.com".
func hostHasLabel(hostname, label string) bool {
	for _, part := range strings.Split(hostname, ".") {
		if part == label {
			return true
		}
	}
	return false
}

// branchPathForHost returns the branch path template for hostname. The user's
// configuration wins over the built-in rules, and DefaultBranchPath applies
// when neither matches. An empty result means the host has no predictable
// branch URL, so only the repository root is safe.
func branchPathForHost(hostname string) string {
	if hostname == "" {
		return ""
	}
	if path, ok := HostBranchPaths[hostname]; ok {
		return path
	}
	if rule, ok := resolveHostRule(hostname); ok {
		return rule.branchPath
	}
	return DefaultBranchPath
}

// getHostingService determines the Git hosting service from the remote URL.
func getHostingService(remoteURL string) HostingService {
	rule, ok := resolveHostRule(hostFromRemoteURL(remoteURL))
	if !ok {
		return Unknown
	}
	return rule.service
}

// buildBranchURL appends the branch path for the remote's host to baseURL.
// When the host is unknown (no built-in rule, no override and no default
// style), baseURL is returned unchanged, so an unknown host opens the
// repository root instead of a guessed URL that would 404.
func buildBranchURL(baseURL, branchName, remoteURL string) string {
	path := branchPathForHost(hostFromRemoteURL(remoteURL))
	if path == "" {
		return baseURL
	}
	return baseURL + strings.ReplaceAll(path, "{branch}", branchName)
}

// stylePath resolves a style name to its branch path template. The second
// result is false when the name is not a known style.
func stylePath(name string) (string, bool) {
	path, ok := branchStyles[strings.ToLower(strings.TrimSpace(name))]
	return path, ok
}

// resolveBranchSpec turns a config value into a branch path template. A value
// containing {branch} is used verbatim as a template; any other value is
// resolved as a style name. An empty value resolves to the repository root.
func resolveBranchSpec(spec string) (string, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return branchPathRoot, true
	}
	if strings.Contains(spec, "{branch}") {
		return spec, true
	}
	return stylePath(spec)
}

// parseDefaultBranchPath resolves the "default_style" config value. It returns
// the repository root path when the value is unset or not a known style.
func parseDefaultBranchPath(spec string) string {
	path, ok := resolveBranchSpec(spec)
	if !ok {
		return branchPathRoot
	}
	return path
}

// parseHostBranchPaths converts the raw "hosts" config value into a map of
// lowercase hostname to resolved branch path template. The value may be a
// style name, a raw template, or an object with a "style" or "branch" key:
//
//	hosts:
//	  git.example.com: gitea
//	  gitlab.example.com: "/-/tree/{branch}"
//	  code.example.com:
//	    style: gitea
//	  legacy.example.com: none
//
// Entries whose host is blank, or whose value cannot be resolved, are ignored.
func parseHostBranchPaths(raw map[string]any) map[string]string {
	paths := make(map[string]string, len(raw))
	for host, value := range raw {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			continue
		}
		spec, ok := hostBranchSpec(value)
		if !ok {
			continue
		}
		path, ok := resolveBranchSpec(spec)
		if !ok {
			continue
		}
		paths[host] = path
	}
	return paths
}

// hostBranchSpec extracts the raw branch spec from a config value. It accepts
// a plain string or an object with a "branch" or "style" key.
func hostBranchSpec(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case map[string]any:
		return branchSpecFromStringMap(v)
	case map[any]any:
		m := make(map[string]any, len(v))
		for key, val := range v {
			if ks, ok := key.(string); ok {
				m[ks] = val
			}
		}
		return branchSpecFromStringMap(m)
	}
	return "", false
}

// branchSpecFromStringMap reads the "branch" or "style" key of a nested config
// object. A raw "branch" template takes precedence over a "style" name.
func branchSpecFromStringMap(v map[string]any) (string, bool) {
	if branch, ok := v["branch"].(string); ok {
		return branch, true
	}
	if style, ok := v["style"].(string); ok {
		return style, true
	}
	return "", false
}
