package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"path/filepath"
)

func Test_rootCmd(t *testing.T) {
	// Setup test repository
	_, cleanup := setupTestRepo(t, "https://github.com/zhaochunqi/git-open.git")
	defer cleanup()

	// Save original openURLInBrowser function
	original := OpenURLInBrowser
	defer func() {
		OpenURLInBrowser = original
	}()

	// Mock openURLInBrowser function
	OpenURLInBrowser = func(url string) error {
		return nil
	}

	tests := []struct {
		name       string
		args       []string
		wantOutput string
	}{
		{
			name:       "with plain flag",
			args:       []string{"--plain"},
			wantOutput: "Web URL: https://github.com/zhaochunqi/git-open\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for testing
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			
			// Set flags
			cmd.Flags().Bool("plain", false, "")
			if err := cmd.Flags().Set("plain", "true"); err != nil {
				t.Fatal(err)
			}

			// Run the root command function
			runE := rootCmd.RunE
			if err := runE(cmd, tt.args); err != nil {
				t.Errorf("rootCmd.RunE() error = %v", err)
				return
			}
			
			if got := buf.String(); got != tt.wantOutput {
				t.Errorf("root command output = %q, want %q", got, tt.wantOutput)
			}
		})
	}
}

func Test_Execute(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		setup     func()
		wantError bool
	}{
		{
			name: "normal execution",
			setup: func() {
				_, cleanup := setupTestRepo(t, "https://github.com/zhaochunqi/git-open.git")
				t.Cleanup(cleanup)
				
				// Mock browser open
				original := OpenURLInBrowser
				OpenURLInBrowser = func(url string) error {
					return nil
				}
				t.Cleanup(func() {
					OpenURLInBrowser = original
				})
			},
			wantError: false,
		},
		{
			name: "no git repo",
			setup: func() {
				// Create and change to temp dir without git repo
				tmpDir, err := os.MkdirTemp("", "no-git")
				if err != nil {
					t.Fatal(err)
				}
				currentDir, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					os.Chdir(currentDir)
					os.RemoveAll(tmpDir)
				})
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			
			if tt.args != nil {
				os.Args = tt.args
			}

			if tt.setup != nil {
				tt.setup()
			}

			if err := Execute(); (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T)
		wantErr bool
	}{
		{
			name: "normal config",
			setup: func(t *testing.T) {
				// Save original home directory
				origHome := os.Getenv("HOME")
				t.Cleanup(func() {
					os.Setenv("HOME", origHome)
				})

				// Create temporary home directory
				tmpHome, err := os.MkdirTemp("", "home")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					os.RemoveAll(tmpHome)
				})

				// Set temporary home directory
				if err := os.Setenv("HOME", tmpHome); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
		},
		{
			name: "with config file",
			setup: func(t *testing.T) {
				// Save original home directory
				origHome := os.Getenv("HOME")
				t.Cleanup(func() {
					os.Setenv("HOME", origHome)
				})

				// Create temporary home directory
				tmpHome, err := os.MkdirTemp("", "home")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					os.RemoveAll(tmpHome)
				})

				// Create .git-open directory
				configDir := filepath.Join(tmpHome, ".git-open")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}

				// Create config file
				configFile := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configFile, []byte("browser: firefox"), 0644); err != nil {
					t.Fatal(err)
				}

				// Set temporary home directory
				if err := os.Setenv("HOME", tmpHome); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			initConfig()
		})
	}
}
