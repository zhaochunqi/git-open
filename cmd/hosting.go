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

// hostRule describes how to link to a branch on one family of hosts.
//
// A rule's branchPath is appended to the repository root URL and may contain
// the placeholder {branch}. An empty branchPath means the host has no
// predictable branch URL, so only the repository root can be opened.
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
	// branchPath is a URL suffix containing {branch}.
	branchPath string
}

// builtinHostRules is the built-in knowledge about hosted Git services.
//
// Matching runs in three tiers: exact host, then host suffix, then keyword.
// Rules are grouped by tier so that, for example, an exact match on
// gitlab.com always wins over the "gitlab" keyword.
var builtinHostRules = []hostRule{
	// Tier 1: exact host matches (public SaaS domains).
	{GitHub, "github.com", "", "/tree/{branch}"},
	{GitLab, "gitlab.com", "", "/-/tree/{branch}"},
	{Bitbucket, "bitbucket.org", "", "/src/{branch}"},
	{Gitea, "codeberg.org", "", "/src/branch/{branch}"},
	{Gitea, "gitea.com", "", "/src/branch/{branch}"},
	{SourceHut, "git.sr.ht", "", "/tree/{branch}"},
	{AzureDevOps, "dev.azure.com", "", "?version=GB{branch}"},

	// Tier 2: host suffix matches (per-tenant domains and deploy variants).
	{AzureDevOps, "visualstudio.com", "", "?version=GB{branch}"},

	// Tier 3: label keyword matches (self-hosted instances).
	{GitLab, "", "gitlab", "/-/tree/{branch}"},
	{Gitea, "", "gitea", "/src/branch/{branch}"},
	{Gitea, "", "forgejo", "/src/branch/{branch}"},
}

// HostBranchPaths holds the user configured host -> branch path overrides,
// loaded from the "hosts" section of the config file. Keys are lowercase
// hostnames. A host mapped to an empty string forces the repository root.
var HostBranchPaths = map[string]string{}

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

// branchPathForHost returns the branch path template for hostname, preferring
// the user's configuration over the built-in rules. An empty string means the
// host has no predictable branch URL, so only the repository root is safe.
func branchPathForHost(hostname string) string {
	if hostname == "" {
		return ""
	}
	if path, ok := HostBranchPaths[hostname]; ok {
		return path
	}
	rule, ok := resolveHostRule(hostname)
	if !ok {
		return ""
	}
	return rule.branchPath
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
// When the host is unknown (no built-in rule and no user override), baseURL is
// returned unchanged, so an unknown host opens the repository root instead of
// a guessed URL that would 404.
func buildBranchURL(baseURL, branchName, remoteURL string) string {
	path := branchPathForHost(hostFromRemoteURL(remoteURL))
	if path == "" {
		return baseURL
	}
	return baseURL + strings.ReplaceAll(path, "{branch}", branchName)
}

// parseHostBranchPaths converts the raw "hosts" config value into a map of
// lowercase hostname to branch path template. It accepts both a plain string
// and a nested object with a "branch" key:
//
//	hosts:
//	  gitlab.example.com: "/-/tree/{branch}"
//	  gitea.example.com:
//	    branch: "/src/branch/{branch}"
//
// Entries whose host is blank, or whose value is neither a string nor an
// object with a string "branch", are ignored.
func parseHostBranchPaths(raw map[string]any) map[string]string {
	paths := make(map[string]string, len(raw))
	for host, value := range raw {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" {
			continue
		}
		if path, ok := hostBranchFromValue(value); ok {
			paths[host] = path
		}
	}
	return paths
}

// hostBranchFromValue extracts a branch path from a config value. It accepts
// a plain string or an object with a string "branch" key. Both
// map[string]any and map[any]any are handled because YAML decoders differ in
// which they produce for nested mappings.
func hostBranchFromValue(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v), true
	case map[string]any:
		if branch, ok := v["branch"].(string); ok {
			return strings.TrimSpace(branch), true
		}
	case map[any]any:
		if branch, ok := v["branch"].(string); ok {
			return strings.TrimSpace(branch), true
		}
	}
	return "", false
}
