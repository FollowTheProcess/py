package cmd

import (
	"fmt"

	"github.com/FollowTheProcess/py/cli/app"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
)

func buildListCmd() *cobra.Command {
	app := app.New()

	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List all found python interpreters.",
		Long: heredoc.Doc(`
		
		List all found python interpreters.
		
		The list command will run py's interpreter finder and
		simply report back the list of interpreters it has found
		and their paths.
		`),
		Example: heredoc.Doc(`
		
		$ py list
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.List(); err != nil {
				return fmt.Errorf("cannot list interpreters: %w", err)
			}
			return nil
		},
	}

	return listCmd
}
