// Package cmd implements the py CLI
package cmd

import (
	"fmt"

	"github.com/FollowTheProcess/py/cli/app"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

var (
	version = "dev" // py version, set at compile time by ldflags
	commit  = ""    // py version's commit hash, set at compile time by ldflags
)

func BuildRootCmd() *cobra.Command {
	app := app.New()

	rootCmd := &cobra.Command{
		Use:           "py [args] [flags]",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Python launcher for Unix.",
		Long: heredoc.Doc(`
		
		Python launcher for Unix.

		Providing a convenient way to launch python 🐍

		py is meant to become your go-to command for launching a python interpreter
		while writing code.

		It does this by trying to find the python interpreter that you most likely
		want to use by looking in a few different places:

		1) Passed version as an argument
		2) An activated virtual environment
		3) A virtual environment in the current or parent directories
		4) The shebang of the target file (if relevant)
		5) The latest version of python on $PATH

		If py reaches the end of the list without finding a valid interpreter,
		it will exit with an error message.
		`),
		Example: heredoc.Doc(`
		
		# Launch the latest version of python on $PATH (or a virtual environment)
		$ py

		# Launch the latest python3 on $PATH
		$ py 3

		# Launch a specific version on $PATH
		$ py 3.10

		# List all found interpreters
		$ py list
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Launch(args); err != nil {
				return fmt.Errorf("error launching: %w", err)
			}
			return nil
		},
	}

	// Attach child commands
	rootCmd.AddCommand(
		buildVersionCmd(),
		buildListCmd(),
	)

	return rootCmd
}
