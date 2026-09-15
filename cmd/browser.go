package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ErrMockBrowser is used for testing browser errors
var ErrMockBrowser = errors.New("mock browser error")

// OpenURLInBrowser is exported for testing
var OpenURLInBrowser = openURLInBrowser

// getPlatform returns the current platform, can be mocked for testing
var getPlatform = func() string {
	return runtime.GOOS
}

// isWSL reports whether the current process runs inside the Windows Subsystem
// for Linux. It can be mocked for testing.
var isWSL = func() bool {
	return detectWSL()
}

// procVersionPath is the kernel version file inspected to detect WSL.
var procVersionPath = "/proc/version"

// lookPath wraps exec.LookPath so tests can stub which commands are installed.
var lookPath = exec.LookPath

var commandRunner = func(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	// Redirect stdout and stderr to /dev/null to suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

var BrowserCommand string

// wslOpeners lists the commands tried, in order, when running under WSL.
// Windows executables are reachable through WSL's PATH interop even though
// they live outside the Linux filesystem, while wslview only exists when the
// wslu package is installed.
var wslOpeners = [][]string{
	{"wslview"},
	{"explorer.exe"},
	{"powershell.exe", "-NoProfile", "-Command", "Start-Process"},
	{"cmd.exe", "/c", "start", ""},
}

// detectWSL reports whether we are running under WSL. WSL sets
// WSL_DISTRO_NAME/WSL_INTEROP, and its kernel version string contains
// "microsoft"; /proc/version is used as a fallback for older versions.
func detectWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}

	data, err := os.ReadFile(procVersionPath)
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

func openURLInBrowser(url string) error {
	platform := getPlatform()

	customBrowser := strings.TrimSpace(BrowserCommand)
	if customBrowser != "" {
		return commandRunner(customBrowser, url)
	}

	// On Linux, use xdg-open with output redirection to suppress messages
	if platform == "linux" {
		// WSL distributions usually have no X server, so xdg-open is either
		// missing or points at nothing; prefer the Windows browser instead.
		if isWSL() {
			return openWithWSL(url)
		}
		return openWithXdgOpen(url)
	}
	// On macOS, use open command with output redirection to suppress messages
	if platform == "darwin" {
		return openWithMacOSOpen(url)
	}
	// On Windows, use start command with output redirection to suppress messages
	if platform == "windows" {
		return openWithWindowsStart(url)
	}
	// For other platforms, return an error
	return errors.New("unsupported platform: " + platform)
}

func openWithXdgOpen(url string) error {
	return commandRunner("xdg-open", url)
}

// openWithWSL opens url under WSL by trying the available openers in order of
// preference. xdg-open is tried first when it exists (WSLg ships one that
// respects the Linux default browser), then wslview, then the Windows
// executables that WSL exposes through PATH interop.
func openWithWSL(url string) error {
	openers := make([][]string, 0, len(wslOpeners)+1)
	if _, err := lookPath("xdg-open"); err == nil {
		openers = append(openers, []string{"xdg-open"})
	}
	openers = append(openers, wslOpeners...)

	var errs []error
	for _, opener := range openers {
		if _, err := lookPath(opener[0]); err != nil {
			continue
		}
		if err := commandRunner(opener[0], wslOpenerArgs(opener, url)...); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", opener[0], err))
			continue
		}
		return nil
	}

	if len(errs) == 0 {
		return errors.New("no browser opener found in $PATH: install wslu (for wslview), enable Windows executable interop in WSL, or set the browser config option")
	}

	return fmt.Errorf("failed to open URL with any WSL browser opener: %w", errors.Join(errs...))
}

// wslOpenerArgs builds the arguments for an opener, appending the URL to the
// opener's fixed arguments. cmd.exe and PowerShell parse the URL themselves, so
// it is passed as a quoted literal to keep shell metacharacters (such as the
// "&" that may appear in a branch name) intact.
func wslOpenerArgs(opener []string, url string) []string {
	args := make([]string, 0, len(opener))
	args = append(args, opener[1:]...)

	switch opener[0] {
	case "powershell.exe":
		url = "'" + strings.ReplaceAll(url, "'", "''") + "'"
	case "cmd.exe":
		url = `"` + url + `"`
	}

	return append(args, url)
}

func openWithMacOSOpen(url string) error {
	return commandRunner("open", url)
}

func openWithWindowsStart(url string) error {
	return commandRunner("cmd", "/c", "start", "", url)
}

func openURLInBrowserFunc(url string) error {
	return OpenURLInBrowser(url)
}
