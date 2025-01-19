package main

import (
	"os"

	"github.com/zhaochunqi/git-open/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
