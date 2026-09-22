# Git Open

[![codecov](https://codecov.io/github/zhaochunqi/git-open/graph/badge.svg?token=TXC9ZOSHFT)](https://codecov.io/github/zhaochunqi/git-open)

**macOS · Linux · Windows · WSL** — a single static binary that opens the current
Git repository in your browser, with no runtime dependencies and no configuration.

> [!IMPORTANT]
> **After installation, just run `git open`:**
>
> ```sh
> git open
> ```
>
> Run it inside any Git repository and that repository opens in your browser —
> on your current branch. That's the whole workflow; `git-open` works too, since
> git treats it as the `open` subcommand. Hosts without a known branch URL still
> open the repository root, and any host can be taught one in the config file.

## 🚀 Quick Install

The fastest way is a package manager:

**mise** (macOS · Linux · WSL):

```sh
mise use -g github:zhaochunqi/git-open
```

**Nix**:

```sh
nix profile install github:zhaochunqi/git-open
```

**Homebrew** (macOS):

```sh
brew install --cask zhaochunqi/tap/git-open
```

Prebuilt binaries are published for every supported platform as well. Pick the asset
that matches your machine and follow [Installation Options](#installation-options):

| Platform | Release asset |
| --- | --- |
| macOS (Intel / Apple Silicon) | `git-open_Darwin_x86_64.tar.gz` / `git-open_Darwin_arm64.tar.gz` |
| Linux (x86_64 / arm64) | `git-open_Linux_x86_64.tar.gz` / `git-open_Linux_arm64.tar.gz` |
| Windows (x86_64 / arm64) | `git-open_Windows_x86_64.zip` / `git-open_Windows_arm64.zip` |
| WSL (WSL1 / WSL2) | install the Linux binary inside your distro — see [WSL](#wsl-windows-subsystem-for-linux) |

## Features

* **Cross-platform by design** — macOS (Intel / Apple Silicon), Linux (x86_64 / arm64)
  and Windows (x86_64 / arm64), plus WSL.
* **WSL aware** — inside WSL the URL is handed to your Windows default browser, so no
  WSLg, X server or `xdg-open` is required.
* **Branch aware** — opens the current branch, while `main` / `master` fall back to the
  repository root.
* **Host aware, vendor neutral** — GitHub, GitLab, Bitbucket, Gitea/Forgejo/Codeberg,
  SourceHut and Azure DevOps are recognised by hostname, and any other host can be added
  in the config file. Unknown hosts open the repository root instead of a guessed URL.
* **Configurable browser** — `~/.config/git-open/config.yaml` (`browser: ...`),
  following the XDG Base Directory spec, or the `BROWSER` environment variable.
* **`git -C` style `-C` flag** — run as if started in another directory.
* **Lightweight** — one static binary per platform and no runtime dependencies.

## Platform Support

| Platform | Status | How the URL is opened |
| --- | --- | --- |
| macOS | supported | `open` |
| Linux | supported | `xdg-open` |
| Windows | supported | `cmd /c start` |
| WSL (WSL1 / WSL2) | supported | `xdg-open` → `wslview` → `explorer.exe` → `powershell.exe` → `cmd.exe` |
| Others (FreeBSD, Plan 9, ...) | not supported | exits with `unsupported platform: <goos>` |

WSL is detected from `WSL_DISTRO_NAME` / `WSL_INTEROP`, with the `/proc/version`
kernel string (`microsoft`) as a fallback. Any platform can be overridden with the
`browser` option — see [Configuration](#configuration) and
[WSL](#wsl-windows-subsystem-for-linux).

## Installation Options

Package managers are the easiest route; use the prebuilt binaries if you would
rather install manually.

### mise

For users who prefer mise for version management (works on macOS, Linux and WSL):

Install the latest release globally:

```sh
mise use -g github:zhaochunqi/git-open
```

Or pin a specific release:

```sh
mise use -g github:zhaochunqi/git-open@2.6.0
```

This uses mise's [`github` backend](https://mise.jdx.dev/dev-tools/backends/github.html), which
installs pre-built binaries straight from GitHub Releases. The older
`ubi:zhaochunqi/git-open` syntax still works but is
[deprecated](https://mise.jdx.dev/dev-tools/backends/ubi.html) — migrate by replacing `ubi:` with
`github:` in your existing config. The equivalent `mise.toml` entry is:

```toml
[tools]
"github:zhaochunqi/git-open" = "latest"
```

### Nix

Using flakes (recommended):

One-shot run without installing:

```sh
nix run github:zhaochunqi/git-open
```

Install into your profile:

```sh
nix profile install github:zhaochunqi/git-open
```

Or pin a release tag:

```sh
nix profile install github:zhaochunqi/git-open/v2.6.0
```

From a local checkout:

```sh
nix build
./result/bin/git-open version
```

Development shell with Go toolchain:

```sh
nix develop
```

### Homebrew (macOS only)

This tap ships a **cask**, and Homebrew casks only work on macOS. (Homebrew itself
also runs on Linux, but without cask support.) On Linux, use mise, Nix or the
prebuilt binary instead.

```sh
brew install --cask zhaochunqi/tap/git-open
```

### Prebuilt binaries (manual, all platforms)

Download the asset for your platform.

**macOS (Intel):**

```sh
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Darwin_x86_64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/
```

**macOS (Apple Silicon):**

```sh
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Darwin_arm64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/
```

**Linux (x86_64):**

```sh
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Linux_x86_64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/
```

**Linux (arm64):**

```sh
curl -L https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Linux_arm64.tar.gz -o git-open.tar.gz
tar -xzf git-open.tar.gz
chmod +x git-open
sudo mv git-open /usr/local/bin/
```

On Windows, download the zip from the same release and add the extracted folder to
your `PATH`:

```powershell
Invoke-WebRequest -Uri https://github.com/zhaochunqi/git-open/releases/latest/download/git-open_Windows_x86_64.zip -OutFile git-open.zip
Expand-Archive -Path git-open.zip -DestinationPath "$env:LOCALAPPDATA\git-open"
# then add %LOCALAPPDATA%\git-open to your PATH
```

Asset names always follow `git-open_<OS>_<arch>`: `Darwin_arm64` for Apple Silicon,
`Linux_arm64` for arm64 Linux, `Windows_arm64` for Windows on ARM.

WSL users install the same Linux binary inside their distribution — there is no
separate package. Everything else is handled automatically; see
[WSL](#wsl-windows-subsystem-for-linux) for the details.

## Usage

Navigate to your project's directory and run the following command:

`git-open` or `git open`

This will open your repository in the default web browser.

To print the repository name (e.g. `github.com/zhaochunqi/git-open`):

`git-open repo`

To run as if started in a different directory (e.g. from a script that isn't inside the repo):

`git-open -C /path/to/repo repo`

The `-C` flag mirrors `git -C`: it may be given multiple times, and a non-absolute path is relative to the previous one.

To print the URL instead of opening a browser:

`git-open --plain` (or `-p`)

## WSL (Windows Subsystem for Linux)

WSL distributions normally have no X server, so `xdg-open` is not installed. When
running under WSL, `git-open` tries these openers in order and uses the first one
available:

| Order | Opener | Notes |
| --- | --- | --- |
| 1 | `xdg-open` | used when WSLg provides it, so your Linux default browser wins |
| 2 | `wslview` | the [wslu](https://github.com/wslutilities/wslu) package, optional |
| 3 | `explorer.exe` | Windows default browser, no extra setup needed |
| 4 | `powershell.exe` | URL passed as a PowerShell single-quoted literal |
| 5 | `cmd.exe` | URL passed as a quoted argument to `start` |

WSL is detected from `WSL_DISTRO_NAME` / `WSL_INTEROP`; when neither is set, the
kernel string in `/proc/version` is checked for `microsoft`, which both WSL1 and WSL2
report. Everything else is ordinary Linux behaviour, so installing the Linux binary
is all the setup required.

If none of the openers is available, `git-open` fails with a pointer to the fix:
install `wslu` (`sudo apt install wslu`), enable Windows executable interop, or
configure a browser explicitly (see [Configuration](#configuration)):

```yaml
browser: wslview
```

A full Windows path works too, which is useful when you want a browser other than the
Windows default one:

```yaml
browser: /mnt/c/Program Files/Google/Chrome/Application/chrome.exe
```

## Configuration

`git-open` reads the first file that exists, in this order (following the
[XDG Base Directory specification](https://specifications.freedesktop.org/basedir-spec/latest/)):

1. the path given to `--config` — always wins when passed
2. `$XDG_CONFIG_HOME/git-open/config.yaml` — i.e. `~/.config/git-open/config.yaml`
   when `XDG_CONFIG_HOME` is not set
3. `~/.git-open.yaml` — the legacy location, still supported

```yaml
# Command used to open the URL, executed with the URL as its only argument.
browser: explorer.exe
```

The `BROWSER` environment variable is honoured as well, so a one-off override needs no
config file:

```sh
BROWSER=wslview git open
```

When `browser` is set it takes precedence over the platform defaults listed in
[Platform Support](#platform-support), which makes it the escape hatch for any
platform-specific behaviour.

### Repository hosts

`git-open` links to the current branch with the path used by the hosting service. The
host is taken from the remote URL, and the built-in rules cover:

| Host | Branch URL suffix |
| --- | --- |
| `github.com` | `/tree/<branch>` |
| `gitlab.com` | `/-/tree/<branch>` |
| `bitbucket.org` | `/src/<branch>` |
| `codeberg.org`, `gitea.com` | `/src/branch/<branch>` |
| `git.sr.ht` | `/tree/<branch>` |
| `dev.azure.com`, `*.visualstudio.com` | `?version=GB<branch>` |
| any hostname with a `gitlab` label | `/-/tree/<branch>` |
| any hostname with a `gitea` or `forgejo` label | `/src/branch/<branch>` |

Self-hosted instances whose hostname does not hint at the software (including GitHub
Enterprise and Azure DevOps Server) are not in the table. For those, and to override a
built-in rule, add a `hosts` section to the config file. The key is the exact hostname
(not the whole remote URL), and the value is the URL suffix appended to the repository
root; `{branch}` is replaced with the branch name:

```yaml
hosts:
  gitlab.example.com: "/-/tree/{branch}"
  gitea.example.com:
    branch: "/src/branch/{branch}"
  github.example.com: "/tree/{branch}"
```

When a host has no rule and no override, `git-open` opens the repository root rather than
guessing a branch URL that would 404. An empty value disables branch links for a host:

```yaml
hosts:
  legacy.example.com: ""
```

## Testing

This project follows Go testing best practices. Here's how to run the tests:

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmark tests
go test -bench=. ./...
```

### Test Structure

The tests are organized as follows:

- Unit tests for core functionality
- Integration tests for git repository operations
- Benchmark tests for performance-critical functions

The test suite uses a custom test utility package (`internal/testhelper`) that provides common testing functions and fixtures.

### Test Coverage

We aim to maintain high test coverage with:
- Multiple test cases for each function
- Edge case testing
- Error condition testing
- Performance benchmarking for critical paths

### Contributing Tests

When adding new features, please ensure:
1. Add corresponding test cases
2. Include both positive and negative test scenarios
3. Add benchmark tests for performance-sensitive functions
4. Use the provided test utilities from `internal/testhelper`

## Contributing

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please submit a pull request. For major changes, please open an issue first to discuss what you would like to change.
