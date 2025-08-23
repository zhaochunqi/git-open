package main

import (
	"fmt"
	"os"

	"github.com/zhaochunqi/git-open/cmd"
)

// exitFunc is a variable that can be replaced for testing os.Exit
var exitFunc = os.Exit

func main() {
	runApp()
}

func runApp() error {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // Print error to stderr
		exitFunc(1)
		return err // Return error for testing purposes
	}
	return nil
}
