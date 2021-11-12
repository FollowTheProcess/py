// Package cmd implements the py CLI
package cmd

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

var (
	version = "dev" // py version, set at compile time by ldflags
	commit  = ""    // py version's commit hash, set at compile time by ldflags
)

func BuildRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "py <command> [flags]",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Short:         "Port of the python-launcher to Go.",
		Long: heredoc.Doc(`
		
		Longer description of your CLI.
		`),
		Example: heredoc.Doc(`

		$ py hello

		$ py version

		$ py --help
		`),
	}

	// Attach child commands
	rootCmd.AddCommand(
		buildVersionCmd(),
		buildHelloCommand(),
	)

	return rootCmd
}
