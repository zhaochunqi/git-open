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
