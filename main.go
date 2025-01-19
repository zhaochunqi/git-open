package main

import (
	"os"
	"testing"

	"github.com/zhaochunqi/git-open/cmd"
)

func TestMain(m *testing.M) {
	cmd.Execute()
	os.Exit(m.Run())
}

func main() {
	cmd.Execute()
}
