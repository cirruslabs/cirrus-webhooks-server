package command

import (
	"github.com/cirruslabs/cirrus-webhooks-server/internal/command/datadog"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cws",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(datadog.NewCommand())

	return cmd
}
