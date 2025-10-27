package cmd

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/zhaochunqi/git-open/internal/testhelper"
	"path/filepath"
)

func Test_rootCmd(t *testing.T) {
	// Setup test repository
	_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
	defer cleanup()

	// Save original openURLInBrowser function and restore after test
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
		name        string
		args        []string
		wantError   bool
		expectedURL string // New field to store the expected URL
	}{
		{
			name:        "normal execution - github main branch",
			wantError:   false,
			expectedURL: "https://github.com/zhaochunqi/git-open",
		},
		{
			name:        "normal execution - github feature branch",
			wantError:   false,
			expectedURL: "https://github.com/zhaochunqi/git-open/tree/feature-branch",
		},
		{
			name:        "normal execution - gitlab feature branch",
			wantError:   false,
			expectedURL: "https://gitlab.com/zhaochunqi/git-open/-/tree/feature-branch",
		},
		{
			name:        "normal execution - bitbucket feature branch",
			wantError:   false,
			expectedURL: "https://bitbucket.org/zhaochunqi/git-open/src/feature-branch",
		},
		{
			name:      "no git repo",
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

			var openedURL string
			originalOpenURLInBrowser := OpenURLInBrowser
			OpenURLInBrowser = func(url string) error {
				openedURL = url
				return nil
			}
			defer func() { OpenURLInBrowser = originalOpenURLInBrowser }()

			// Setup logic moved directly into the test case
			switch tt.name {
			case "normal execution - github main branch":
				_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "main")
				t.Cleanup(cleanup)
			case "normal execution - github feature branch":
				_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/zhaochunqi/git-open.git", "feature-branch")
				t.Cleanup(cleanup)
			case "normal execution - gitlab feature branch":
				_, cleanup := testhelper.SetupTestRepo(t, "https://gitlab.com/zhaochunqi/git-open.git", "feature-branch")
				t.Cleanup(cleanup)
			case "normal execution - bitbucket feature branch":
				_, cleanup := testhelper.SetupTestRepo(t, "https://bitbucket.org/zhaochunqi/git-open.git", "feature-branch")
				t.Cleanup(cleanup)
			case "no git repo":
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
			}

			if err := Execute(); (err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && tt.expectedURL != "" && openedURL != tt.expectedURL {
				t.Errorf("Execute() opened URL = %q, want %q", openedURL, tt.expectedURL)
			}
		})
	}
}

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "normal config",
			wantErr: false,
		},
		{
			name:    "with config file",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logic moved directly into the test case
			switch tt.name {
			case "normal config":
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
			case "with config file":
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
			}

			initConfig()
		})
	}
}
func Test_rootCmd_ErrorHandling(t *testing.T) {
	// Save original functions
	originalGetCurrentGitDirectoryFunc := getCurrentGitDirectoryFunc
	originalGetRemoteURLFunc := getRemoteURLFunc
	originalGetBranchNameFunc := getBranchNameFunc
	originalOpenURLInBrowser := OpenURLInBrowser

	defer func() {
		getCurrentGitDirectoryFunc = originalGetCurrentGitDirectoryFunc
		getRemoteURLFunc = originalGetRemoteURLFunc
		getBranchNameFunc = originalGetBranchNameFunc
		OpenURLInBrowser = originalOpenURLInBrowser
	}()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "error getting git directory",
			setup: func() {
				getCurrentGitDirectoryFunc = func() (*git.Repository, error) {
					return nil, errors.New("git directory error")
				}
			},
			wantErr: true,
		},
		{
			name: "error getting remote URL",
			setup: func() {
				_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
				t.Cleanup(cleanup)

				getRemoteURLFunc = func(repo *git.Repository) (string, error) {
					return "", errors.New("remote URL error")
				}
			},
			wantErr: true,
		},
		{
			name: "error getting branch name",
			setup: func() {
				_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
				t.Cleanup(cleanup)

				getBranchNameFunc = func(repo *git.Repository) (string, error) {
					return "", errors.New("branch name error")
				}
			},
			wantErr: true,
		},
		{
			name: "browser error",
			setup: func() {
				_, cleanup := testhelper.SetupTestRepo(t, "https://github.com/test/repo.git", "main")
				t.Cleanup(cleanup)

				OpenURLInBrowser = func(url string) error {
					return errors.New("browser error")
				}
			},
			wantErr: false, // Browser error doesn't cause command to fail, it just prints error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset functions to original before each test
			getCurrentGitDirectoryFunc = originalGetCurrentGitDirectoryFunc
			getRemoteURLFunc = originalGetRemoteURLFunc
			getBranchNameFunc = originalGetBranchNameFunc
			OpenURLInBrowser = originalOpenURLInBrowser

			tt.setup()

			// Create a buffer to capture stderr
			buf := new(bytes.Buffer)
			cmd := &cobra.Command{}
			cmd.SetErr(buf)

			// Run the root command function
			runE := rootCmd.RunE
			err := runE(cmd, []string{})

			if (err != nil) != tt.wantErr {
				t.Errorf("rootCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initConfig_ErrorCases(t *testing.T) {
	// Save original cfgFile
	originalCfgFile := cfgFile
	defer func() {
		cfgFile = originalCfgFile
	}()

	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "with specific config file",
			setup: func() {
				// Create a temporary config file
				tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
				if err != nil {
					t.Fatal(err)
				}
				tmpFile.WriteString("test: value\n")
				tmpFile.Close()
				t.Cleanup(func() {
					os.Remove(tmpFile.Name())
				})

				cfgFile = tmpFile.Name()
			},
		},
		{
			name: "config file read error",
			setup: func() {
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

				cfgFile = ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			initConfig()
		})
	}
}
