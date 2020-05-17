package cfn

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cfn",
		Short: "commands for CloudFormation operation",
		Long:  `About usage, check each sub commands help`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}
	return cmd
}
