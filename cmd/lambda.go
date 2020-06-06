package cmd

import (
	"github.com/Blue-Pix/abc/lib/lambda"
	"github.com/Blue-Pix/abc/lib/lambda/stats"
)

var lambdaCmd = lambda.NewCmd()
var statsCmd = stats.NewCmd()

func init() {
	lambdaCmd.SetOut(rootCmd.OutOrStdout())
	rootCmd.AddCommand(lambdaCmd)
	lambdaCmd.AddCommand(statsCmd)
}
