package cmd

import (
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/unusedExports"
)

var cfnCmd = cfn.NewCmd()
var unusedExportsCmd = unusedExports.NewCmd()

func init() {
	cfnCmd.SetOut(rootCmd.OutOrStdout())
	rootCmd.AddCommand(cfnCmd)
	cfnCmd.AddCommand(unusedExportsCmd)
}
