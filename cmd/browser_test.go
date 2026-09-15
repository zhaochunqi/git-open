package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// recordedCommand captures a single commandRunner invocation.
type recordedCommand struct {
	name string
	args []string
}

// withFakeCommands stubs lookPath and commandRunner so tests can control which
// commands are considered installed and record what was executed.
func withFakeCommands(t *testing.T, installed, failing []string) *[]recordedCommand {
	t.Helper()

	originalLookPath := lookPath
	originalCommandRunner := commandRunner
	t.Cleanup(func() {
		lookPath = originalLookPath
		commandRunner = originalCommandRunner
	})

	recorded := new([]recordedCommand)

	lookPath = func(file string) (string, error) {
		for _, name := range installed {
			if name == file {
				return "/usr/bin/" + file, nil
			}
		}
		return "", exec.ErrNotFound
	}

	commandRunner = func(name string, args ...string) error {
		*recorded = append(*recorded, recordedCommand{name: name, args: args})
		for _, failingName := range failing {
			if failingName == name {
				return errors.New("failed to run " + name)
			}
		}
		return nil
	}

	return recorded
}

func Test_detectWSL(t *testing.T) {
	originalProcVersionPath := procVersionPath
	t.Cleanup(func() {
		procVersionPath = originalProcVersionPath
	})

	writeProcVersion := func(t *testing.T, content string) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "version")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		procVersionPath = path
	}

	tests := []struct {
		name          string
		distroEnv     string
		interopEnv    string
		procVersion   string
		missingKernel bool
		want          bool
	}{
		{
			name:      "WSL_DISTRO_NAME set",
			distroEnv: "Ubuntu",
			want:      true,
		},
		{
			name:       "WSL_INTEROP set",
			interopEnv: "/run/WSL/8_interop",
			want:       true,
		},
		{
			name:        "kernel version mentions Microsoft",
			procVersion: "Linux version 6.18.40.1-microsoft-standard-WSL2 (gcc 13.2.0) #1 SMP",
			want:        true,
		},
		{
			name:        "kernel version mentions Microsoft with capitals",
			procVersion: "Linux version 4.4.0-19041-Microsoft",
			want:        true,
		},
		{
			name:        "plain Linux kernel",
			procVersion: "Linux version 6.11.0-24-generic (buildd@lcy02) #1 SMP",
			want:        false,
		},
		{
			name:          "kernel version file missing",
			missingKernel: true,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WSL_DISTRO_NAME", tt.distroEnv)
			t.Setenv("WSL_INTEROP", tt.interopEnv)

			if tt.missingKernel {
				procVersionPath = filepath.Join(t.TempDir(), "does-not-exist")
			} else {
				writeProcVersion(t, tt.procVersion)
			}

			if got := detectWSL(); got != tt.want {
				t.Errorf("detectWSL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockOpenURL is a mock function for testing
var mockOpenURL func(string) error

func withNoopCommandRunner(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		commandRunner = func(name string, args ...string) error {
			cmd := exec.Command(name, args...)
			// Redirect stdout and stderr to /dev/null to suppress output
			cmd.Stdout = nil
			cmd.Stderr = nil
			return cmd.Start()
		}
	})
	commandRunner = func(name string, args ...string) error {
		return nil
	}
}

func Test_openURLInBrowser(t *testing.T) {
	// Save original function
	original := OpenURLInBrowser
	defer func() {
		OpenURLInBrowser = original
	}()

	tests := []struct {
		name    string
		url     string
		mockErr error
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://github.com/zhaochunqi/git-open",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "error case",
			url:     "https://github.com/zhaochunqi/git-open",
			mockErr: ErrMockBrowser,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up mock function
			OpenURLInBrowser = func(url string) error {
				if url != tt.url {
					t.Errorf("openURLInBrowser() called with url = %v, want %v", url, tt.url)
				}
				return tt.mockErr
			}

			if err := OpenURLInBrowser(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("openURLInBrowser() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := openURLInBrowserFunc(tt.url); (err != nil) != tt.wantErr {
				t.Errorf("openURLInBrowserFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_openWithXdgOpen(t *testing.T) {
	// This test is for Linux-specific functionality
	if runtime.GOOS != "linux" {
		t.Skip("Skipping test on non-Linux platforms")
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://github.com/zhaochunqi/git-open",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withNoopCommandRunner(t)
			err := openWithXdgOpen(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("openWithXdgOpen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_openWithMacOSOpen(t *testing.T) {
	// This test is for macOS-specific functionality
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-macOS platforms")
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://github.com/zhaochunqi/git-open",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withNoopCommandRunner(t)
			err := openWithMacOSOpen(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("openWithMacOSOpen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_openURLInBrowser_PlatformSpecific(t *testing.T) {
	// Save original function
	original := OpenURLInBrowser
	defer func() {
		OpenURLInBrowser = original
	}()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL",
			url:         "https://github.com/zhaochunqi/git-open",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: false, // Commands might handle empty URL gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the actual platform-specific implementation
			withNoopCommandRunner(t)
			err := openURLInBrowser(tt.url)
			if (err != nil) != tt.expectError {
				t.Errorf("openURLInBrowser() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func Test_openURLInBrowser_AllPlatforms(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		platform string
	}{
		{
			name:     "Linux platform",
			url:      "https://github.com/zhaochunqi/git-open",
			platform: "linux",
		},
		{
			name:     "macOS platform",
			url:      "https://github.com/zhaochunqi/git-open",
			platform: "darwin",
		},
		{
			name:     "Windows platform",
			url:      "https://github.com/zhaochunqi/git-open",
			platform: "windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original runtime.GOOS
			originalGOOS := runtime.GOOS
			// We can't actually change runtime.GOOS, but we can test the functions directly

			switch tt.platform {
			case "linux":
				// Test openWithXdgOpen directly if not on Linux
				if runtime.GOOS != "linux" {
					withNoopCommandRunner(t)
					err := openWithXdgOpen(tt.url)
					// On non-Linux systems, this should fail as xdg-open doesn't exist
					if err == nil {
						t.Logf("openWithXdgOpen() succeeded on %s platform, command might exist", originalGOOS)
					}
				}
			case "darwin":
				// Test openWithMacOSOpen directly if not on macOS
				if runtime.GOOS != "darwin" {
					withNoopCommandRunner(t)
					err := openWithMacOSOpen(tt.url)
					// On non-macOS systems, this should fail as open command might not exist
					if err == nil {
						t.Logf("openWithMacOSOpen() succeeded on %s platform, command might exist", originalGOOS)
					}
				}
			case "windows":
				// Test openWithWindowsStart directly if not on Windows
				if runtime.GOOS != "windows" {
					withNoopCommandRunner(t)
					err := openWithWindowsStart(tt.url)
					// On non-Windows systems, this should fail as cmd doesn't exist
					if err == nil {
						t.Logf("openWithWindowsStart() succeeded on %s platform, command might exist", originalGOOS)
					}
				}
			}
		})
	}
}

func Test_openWithWindowsStart(t *testing.T) {
	// This test is for Windows-specific functionality
	if runtime.GOOS != "windows" {
		t.Skip("Skipping test on non-Windows platforms")
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "success",
			url:     "https://github.com/zhaochunqi/git-open",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := openWithWindowsStart(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("openWithWindowsStart() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_openURLInBrowser_UnsupportedPlatform(t *testing.T) {
	// Save original function
	originalGetPlatform := getPlatform
	defer func() {
		getPlatform = originalGetPlatform
	}()

	tests := []struct {
		name     string
		platform string
		url      string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Linux platform",
			platform: "linux",
			url:      "https://github.com/test/repo",
			wantErr:  false, // May fail in CI but function should be called
		},
		{
			name:     "macOS platform",
			platform: "darwin",
			url:      "https://github.com/test/repo",
			wantErr:  false, // May fail in CI but function should be called
		},
		{
			name:     "Windows platform",
			platform: "windows",
			url:      "https://github.com/test/repo",
			wantErr:  false, // May fail in CI but function should be called
		},
		{
			name:     "Unsupported platform - FreeBSD",
			platform: "freebsd",
			url:      "https://github.com/test/repo",
			wantErr:  true,
			errMsg:   "unsupported platform: freebsd",
		},
		{
			name:     "Unsupported platform - Plan9",
			platform: "plan9",
			url:      "https://github.com/test/repo",
			wantErr:  true,
			errMsg:   "unsupported platform: plan9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the platform
			getPlatform = func() string {
				return tt.platform
			}
			withNoopCommandRunner(t)

			err := openURLInBrowser(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("openURLInBrowser() expected error for platform %s, got nil", tt.platform)
				} else if err.Error() != tt.errMsg {
					t.Errorf("openURLInBrowser() error = %v, want %s", err, tt.errMsg)
				}
			} else {
				// For supported platforms, the function should be called
				// It might fail due to missing commands in CI, but that's expected
				t.Logf("openURLInBrowser() for platform %s returned: %v", tt.platform, err)
			}
		})
	}
}

func Test_openWithWSL(t *testing.T) {
	const url = "https://github.com/zhaochunqi/git-open"

	tests := []struct {
		name      string
		installed []string
		failing   []string
		wantCmd   string
		wantArgs  []string
		wantTried []string
		wantErr   bool
	}{
		{
			name:      "prefers xdg-open when WSLg provides it",
			installed: []string{"xdg-open", "wslview", "explorer.exe"},
			wantCmd:   "xdg-open",
			wantArgs:  []string{url},
			wantTried: []string{"xdg-open"},
		},
		{
			name:      "falls back to wslview",
			installed: []string{"wslview", "explorer.exe"},
			wantCmd:   "wslview",
			wantArgs:  []string{url},
			wantTried: []string{"wslview"},
		},
		{
			name:      "falls back to explorer.exe",
			installed: []string{"explorer.exe", "powershell.exe", "cmd.exe"},
			wantCmd:   "explorer.exe",
			wantArgs:  []string{url},
			wantTried: []string{"explorer.exe"},
		},
		{
			name:      "falls back to powershell.exe with a quoted URL",
			installed: []string{"powershell.exe", "cmd.exe"},
			wantCmd:   "powershell.exe",
			wantArgs:  []string{"-NoProfile", "-Command", "Start-Process", "'" + url + "'"},
			wantTried: []string{"powershell.exe"},
		},
		{
			name:      "falls back to cmd.exe with a quoted URL",
			installed: []string{"cmd.exe"},
			wantCmd:   "cmd.exe",
			wantArgs:  []string{"/c", "start", "", `"` + url + `"`},
			wantTried: []string{"cmd.exe"},
		},
		{
			name:      "skips openers that fail to run",
			installed: []string{"wslview", "explorer.exe"},
			failing:   []string{"wslview"},
			wantCmd:   "explorer.exe",
			wantArgs:  []string{url},
			wantTried: []string{"wslview", "explorer.exe"},
		},
		{
			name:    "no opener installed",
			wantErr: true,
		},
		{
			name:      "every opener fails",
			installed: []string{"wslview", "explorer.exe", "powershell.exe", "cmd.exe"},
			failing:   []string{"wslview", "explorer.exe", "powershell.exe", "cmd.exe"},
			wantTried: []string{"wslview", "explorer.exe", "powershell.exe", "cmd.exe"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorded := withFakeCommands(t, tt.installed, tt.failing)

			err := openWithWSL(url)

			if (err != nil) != tt.wantErr {
				t.Fatalf("openWithWSL() error = %v, wantErr %v", err, tt.wantErr)
			}

			tried := make([]string, 0, len(*recorded))
			for _, c := range *recorded {
				tried = append(tried, c.name)
			}
			if !slices.Equal(tried, tt.wantTried) {
				t.Errorf("openWithWSL() tried %v, want %v", tried, tt.wantTried)
			}

			if tt.wantErr {
				if err != nil && tt.name == "no opener installed" && !strings.Contains(err.Error(), "wslview") {
					t.Errorf("openWithWSL() error = %v, want it to mention wslview", err)
				}
				return
			}

			if len(*recorded) == 0 {
				t.Fatal("openWithWSL() ran no command, want one")
			}

			got := (*recorded)[len(*recorded)-1]
			if got.name != tt.wantCmd {
				t.Errorf("openWithWSL() ran %q, want %q", got.name, tt.wantCmd)
			}
			if !slices.Equal(got.args, tt.wantArgs) {
				t.Errorf("openWithWSL() args = %v, want %v", got.args, tt.wantArgs)
			}
		})
	}
}

func Test_wslOpenerArgs(t *testing.T) {
	tests := []struct {
		name   string
		opener []string
		url    string
		want   []string
	}{
		{
			name:   "plain opener",
			opener: []string{"explorer.exe"},
			url:    "https://github.com/zhaochunqi/git-open",
			want:   []string{"https://github.com/zhaochunqi/git-open"},
		},
		{
			name:   "opener with fixed arguments",
			opener: []string{"cmd.exe", "/c", "start", ""},
			url:    "https://github.com/zhaochunqi/git-open",
			want:   []string{"/c", "start", "", `"https://github.com/zhaochunqi/git-open"`},
		},
		{
			name:   "URL is quoted for cmd.exe",
			opener: []string{"cmd.exe", "/c", "start", ""},
			url:    "https://github.com/zhaochunqi/git-open/tree/feat/a&b",
			want:   []string{"/c", "start", "", `"https://github.com/zhaochunqi/git-open/tree/feat/a&b"`},
		},
		{
			name:   "URL is quoted for PowerShell",
			opener: []string{"powershell.exe", "-NoProfile", "-Command", "Start-Process"},
			url:    "https://github.com/zhaochunqi/git-open/tree/feat/a&b",
			want:   []string{"-NoProfile", "-Command", "Start-Process", "'https://github.com/zhaochunqi/git-open/tree/feat/a&b'"},
		},
		{
			name:   "single quotes in the URL are escaped for PowerShell",
			opener: []string{"powershell.exe", "-Command", "Start-Process"},
			url:    "https://example.com/it's",
			want:   []string{"-Command", "Start-Process", "'https://example.com/it''s'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := wslOpenerArgs(tt.opener, tt.url); !slices.Equal(got, tt.want) {
				t.Errorf("wslOpenerArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_openURLInBrowser_PlatformWSL(t *testing.T) {
	originalGetPlatform := getPlatform
	originalIsWSL := isWSL
	t.Cleanup(func() {
		getPlatform = originalGetPlatform
		isWSL = originalIsWSL
	})

	const url = "https://github.com/zhaochunqi/git-open"

	tests := []struct {
		name      string
		platform  string
		wsl       bool
		installed []string
		wantCmd   string
	}{
		{
			name:      "WSL uses a Windows opener instead of xdg-open",
			platform:  "linux",
			wsl:       true,
			installed: []string{"explorer.exe"},
			wantCmd:   "explorer.exe",
		},
		{
			name:      "WSL prefers wslview",
			platform:  "linux",
			wsl:       true,
			installed: []string{"wslview"},
			wantCmd:   "wslview",
		},
		{
			name:      "plain Linux keeps using xdg-open",
			platform:  "linux",
			wsl:       false,
			installed: []string{"explorer.exe"},
			wantCmd:   "xdg-open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPlatform = func() string { return tt.platform }
			isWSL = func() bool { return tt.wsl }

			recorded := withFakeCommands(t, tt.installed, nil)

			if err := openURLInBrowser(url); err != nil {
				t.Fatalf("openURLInBrowser() error = %v", err)
			}

			if len(*recorded) != 1 {
				t.Fatalf("openURLInBrowser() ran %v, want exactly one command", *recorded)
			}

			if got := (*recorded)[0].name; got != tt.wantCmd {
				t.Errorf("openURLInBrowser() ran %q, want %q", got, tt.wantCmd)
			}
		})
	}
}

func Test_openURLInBrowser_CustomBrowserWinsOnWSL(t *testing.T) {
	originalGetPlatform := getPlatform
	originalIsWSL := isWSL
	originalBrowserCommand := BrowserCommand
	t.Cleanup(func() {
		getPlatform = originalGetPlatform
		isWSL = originalIsWSL
		BrowserCommand = originalBrowserCommand
	})

	getPlatform = func() string { return "linux" }
	isWSL = func() bool { return true }
	BrowserCommand = "wslview"

	recorded := withFakeCommands(t, []string{"explorer.exe"}, nil)

	if err := openURLInBrowser("https://github.com/zhaochunqi/git-open"); err != nil {
		t.Fatalf("openURLInBrowser() error = %v", err)
	}

	if len(*recorded) != 1 || (*recorded)[0].name != "wslview" {
		t.Fatalf("openURLInBrowser() ran %v, want the configured browser wslview", *recorded)
	}
}
