package network

import (
	"github.com/spf13/cobra"

	"github.com/crazy-max/docker-cli/cli"
	"github.com/crazy-max/docker-cli/cli/command"
)

// NewNetworkCommand returns a cobra command for `network` subcommands
func NewNetworkCommand(dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "network",
		Short:       "Manage networks",
		Args:        cli.NoArgs,
		RunE:        command.ShowHelp(dockerCli.Err()),
		Annotations: map[string]string{"version": "1.21"},
	}
	cmd.AddCommand(
		newConnectCommand(dockerCli),
		newCreateCommand(dockerCli),
		newDisconnectCommand(dockerCli),
		newInspectCommand(dockerCli),
		newListCommand(dockerCli),
		newRemoveCommand(dockerCli),
		NewPruneCommand(dockerCli),
	)
	return cmd
}
