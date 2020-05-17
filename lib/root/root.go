package root

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "abc",
		Short:         "helper command to become friends with AWS🤝",
		Long:          `A usage, please read each sub commands`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()
			return nil
		},
	}
	// aws profile
	cmd.PersistentFlags().String("profile", "", "(optional) which aws profile to use")
	return cmd
}
