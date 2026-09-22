# ADR 0001: Vendor-neutral host rules for branch links

- **Status:** Accepted
- **Date:** 2025-09-22
- **Deciders:** git-open maintainers

## Context

git-open advertises that it can be run *inside any Git repository* and open
that repository on the current branch. There are two distinct capabilities
behind that promise:

1. **Open the repository root.** This is already vendor-neutral:
   `convertToWebURL` only normalises the scheme and the `.git` suffix and never
   inspects the hostname, so any remote produces a correct root URL.

2. **Open the current branch.** This was vendor-locked. `getHostingService`
   matched the *entire remote URL* with `strings.Contains` against exactly
   three literals (`github.com`, `gitlab.com`, `bitbucket.org`), and
   `buildBranchURL` mapped the result to a hardcoded path, defaulting to the
   GitHub shape (`/tree/{branch}`) for everything else.

That design had two concrete failure modes:

- **Wrong URLs for common hosts.** Self-hosted GitLab
  (`gitlab.example.com`) needs `/-/tree/{branch}` but received `/tree/{branch}`;
  Codeberg/Gitea/Forgejo need `/src/branch/{branch}`; Azure DevOps builds
  query-string URLs. All of these silently produced 404s.
- **False positives.** Matching on the whole URL meant a remote such as
  `https://git.example.com/mirror/github.com.git` was classified as GitHub, and
  a hostname such as `notgitlab.com` matched the `gitlab.com` substring.

The GitHub Enterprise fallback only worked by accident: `/tree/{branch}` is
correct for GHE, so the "default to GitHub" branch hid the problem for one
class of self-hosted users while breaking every other class.

## Decision

Make branch linking data-driven and vendor-neutral.

1. **Resolve the host first.** `hostFromRemoteURL` extracts the lowercase
   hostname from both URL-style (`https://`, `ssh://`, `git+ssh://`) and
   scp-style (`git@host:path`) remotes. Classification and path lookup operate
   on the hostname only, never on the full URL.

2. **Match hosts against a rule table** (`builtinHostRules`) in three tiers:
   exact host, dot-delimited suffix, then a whole hostname label (`gitlab`
   matches `gitlab.example.com` and `code.gitlab.example.com` but not
   `notgitlab.com`). Each rule carries a `branchPath` template containing
   `{branch}`.

3. **Never guess for unknown hosts.** When no rule and no override match, the
   repository root URL is opened. A wrong branch URL is worse than no branch
   link.

4. **Let users extend the table** through the `hosts` section of the config
   file, accepting either a plain string or a nested `branch:` key. An empty
   value forces the repository root for that host.

```yaml
hosts:
  gitlab.example.com: "/-/tree/{branch}"
  gitea.example.com:
    branch: "/src/branch/{branch}"
  github.mycorp.com: "/tree/{branch}"
```

The built-in rules are:

| Host pattern | Service | Branch path |
| --- | --- | --- |
| `github.com` | GitHub | `/tree/{branch}` |
| `gitlab.com` | GitLab | `/-/tree/{branch}` |
| `bitbucket.org` | Bitbucket | `/src/{branch}` |
| `codeberg.org`, `gitea.com` | Gitea | `/src/branch/{branch}` |
| `git.sr.ht` | SourceHut | `/tree/{branch}` |
| `dev.azure.com`, `*.visualstudio.com` | Azure DevOps | `?version=GB{branch}` |
| any label `gitlab` | GitLab | `/-/tree/{branch}` |
| any label `gitea`, `forgejo` | Gitea | `/src/branch/{branch}` |

## Consequences

- **Behaviour change:** unknown hosts no longer receive a `/tree/{branch}`
  URL. Repositories on GitHub Enterprise or other self-hosted instances
  reachable only by a custom domain now open the repository root; users who
  want branch links configure them under `hosts`. This is intentional and
  documented.
- **Extensible without a release:** a new hosting provider can be supported for
  one user, or for everyone, by editing the rule table or the config file.
- **Small new config surface:** `hosts` must be documented and validated. The
  parser ignores malformed entries rather than failing the command.
- **The hostname is the only contract:** providers that multiplex Git repositories
  under a generic host (for example a monorepo gateway) still need a user
  override, since no static rule can identify them.

## Alternatives considered

- **Keep defaulting to `/tree/{branch}`.** Rejected: it silently produces 404s
  for self-hosted GitLab, Gitea and others, and contradicts the "any
  repository" promise.
- **Hardcode a longer list of public domains.** Rejected as the only mechanism:
  it cannot cover the long tail of self-hosted instances, and it does not
  address false-positive matching.
- **Detect the provider over the network (API probe, `/-/` probe).** Rejected:
  it adds latency, requires connectivity, and breaks the "no runtime
  dependencies / offline" property.
- **Infer from `git remote -v` or tracking refs.** Rejected: it does not carry
  provider information beyond the URL already used.

## References

- Prior art: `paulirish/git-open` maintains a comparable host-to-URL table.
- Implementation: `cmd/hosting.go`.
- Tests: `cmd/hosting_test.go`.
