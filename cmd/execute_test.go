package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

// Test_ExecuteWithError tests the behavior of Execute function when rootCmd.Execute returns an error
func Test_ExecuteWithError(t *testing.T) {
	// Create a temporary command to simulate an error condition
	tmpCmd := &cobra.Command{
		Use: "test-error",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("mock command error")
		},
	}

	// Save the original rootCmd
	originalRootCmd := rootCmd
	// Temporarily replace rootCmd
	rootCmd = tmpCmd
	// Restore the original rootCmd after the test
	defer func() {
		rootCmd = originalRootCmd
	}()

	// Execute the function and check if it returns an error
	err := Execute()
	if err == nil {
		t.Error("Execute() should return error when rootCmd.Execute fails")
	}

	// Verify the error message
	if err != nil && err.Error() != "mock command error" {
		t.Errorf("Expected error message 'mock command error', got '%s'", err.Error())
	}
}
