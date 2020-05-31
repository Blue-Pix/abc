package cmd

import (
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/purge_stack"
	"github.com/Blue-Pix/abc/lib/cfn/unused_exports"
)

var cfnCmd = cfn.NewCmd()
var unusedExportsCmd = unused_exports.NewCmd()
var purgeStackCmd = purge_stack.NewCmd()

func init() {
	cfnCmd.SetOut(rootCmd.OutOrStdout())
	rootCmd.AddCommand(cfnCmd)
	cfnCmd.AddCommand(unusedExportsCmd)
	cfnCmd.AddCommand(purgeStackCmd)
}
