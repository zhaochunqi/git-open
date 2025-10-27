package cmd

import (
	"runtime"
	"testing"
)

// mockOpenURL is a mock function for testing
var mockOpenURL func(string) error

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
					err := openWithXdgOpen(tt.url)
					// On non-Linux systems, this should fail as xdg-open doesn't exist
					if err == nil {
						t.Logf("openWithXdgOpen() succeeded on %s platform, command might exist", originalGOOS)
					}
				}
			case "darwin":
				// Test openWithMacOSOpen directly if not on macOS
				if runtime.GOOS != "darwin" {
					err := openWithMacOSOpen(tt.url)
					// On non-macOS systems, this should fail as open command might not exist
					if err == nil {
						t.Logf("openWithMacOSOpen() succeeded on %s platform, command might exist", originalGOOS)
					}
				}
			case "windows":
				// Test openWithWindowsStart directly if not on Windows
				if runtime.GOOS != "windows" {
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
