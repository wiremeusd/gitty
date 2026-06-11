package main

import (
	"github.com/spf13/cobra"
)

// version is injected by GoReleaser via ldflags.
var version = "dev"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "gitty",
		Short:         "Interactively clone GitHub repositories",
		Long:          "gitty shows the repository list for the GitHub account bound to the current folder.\nNavigate with arrow keys, search with /, press Enter to clone into the current folder.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runInteractive,
	}
	root.AddCommand(newAuthCmd(), newUseCmd(), newCredentialCmd())
	return root
}
