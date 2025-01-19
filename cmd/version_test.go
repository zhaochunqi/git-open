package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
)

func Test_versionCmd(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommitHash := CommitHash
	origBuildDate := BuildDate
	
	// Restore original values after test
	defer func() {
		Version = origVersion
		CommitHash = origCommitHash
		BuildDate = origBuildDate
	}()
	
	// Set test values
	Version = "1.0.0"
	CommitHash = "abc123"
	BuildDate = "2025-01-19"
	
	// Create buffer to capture output
	buf := new(bytes.Buffer)
	cmd := &cobra.Command{}
	cmd.SetOut(buf)

	// Run the version command function
	versionCmd.Run(cmd, []string{})

	expected := fmt.Sprintf("Version: %s\nGit Commit: %s\nBuild Date: %s\n",
		Version, CommitHash, BuildDate)
	
	if got := buf.String(); got != expected {
		t.Errorf("version command output = %q, want %q", got, expected)
	}
}
