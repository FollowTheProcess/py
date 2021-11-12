package main

import (
	"os"

	"github.com/FollowTheProcess/msg"
	"github.com/FollowTheProcess/py/cli/cmd"
)

func main() {
	rootCmd := cmd.BuildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		prefix := msg.Sfail("Error:")
		msg.Textf("%s %s", prefix, err)
		os.Exit(1)
	}
}
