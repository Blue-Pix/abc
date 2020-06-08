package stats

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var LambdaClient lambdaiface.LambdaAPI

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show stats of Lambda functions",
		Long: `
[abc lambda stats]
This command describes Lambda functions count by runtime.
By default, output format is markdown table.

Internally it uses aws lambda api.
Please configure your aws credentials with following policies.
- lambda:ListFunctions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	data, err := FetchData(cmd, args)
	if err != nil {
		return err
	}
	var str string
	if len(data) == 0 {
		str = "no function found"
	} else {
		str = Output(data)
	}
	cmd.Print(str)
	return nil
}

func FetchData(cmd *cobra.Command, args []string) (map[string][]string, error) {
	initClient(cmd)
	functions, err := listFunctions()
	if err != nil {
		return nil, err
	}
	count := countByRuntime(functions)
	return count, nil
}

func initClient(cmd *cobra.Command) {
	profile, _ := cmd.Flags().GetString("profile")
	region, _ := cmd.Flags().GetString("region")
	if LambdaClient == nil {
		sess := util.CreateSession(profile, region)
		LambdaClient = lambda.New(sess)
	}
}

func listFunctions() ([]*lambda.FunctionConfiguration, error) {
	var result []*lambda.FunctionConfiguration
	var nextMarker *string
	for {
		params := &lambda.ListFunctionsInput{
			MaxItems: aws.Int64(1000),
			Marker:   nextMarker,
		}
		resp, err := LambdaClient.ListFunctions(params)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.Functions...)
		if resp.NextMarker == nil {
			break
		}
		nextMarker = resp.NextMarker
	}

	return result, nil
}

func countByRuntime(result []*lambda.FunctionConfiguration) map[string][]string {
	m := make(map[string][]string)
	for _, f := range result {
		runtime := aws.StringValue(f.Runtime)
		if _, hasKey := m[runtime]; hasKey == false {
			m[runtime] = []string{aws.StringValue(f.FunctionName)}
		} else {
			m[runtime] = append(m[runtime], aws.StringValue(f.FunctionName))
		}
	}
	return m
}

func sortKey(m map[string][]string) []string {
	keys := make([]string, len(m))
	index := 0
	for key := range m {
		keys[index] = key
		index++
	}
	sort.Strings(keys)
	return keys
}

func Output(count map[string][]string) string {
	keys := sortKey(count)
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"Runtime", "Count"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, k := range keys {
		if isDeprecatedRuntime(k) {
			table.Append([]string{fmt.Sprintf("\x1b[91m%s（Deprecated）\x1b[0m", k), strconv.Itoa(len(count[k]))})
		} else {
			table.Append([]string{k, strconv.Itoa(len(count[k]))})
		}
	}
	table.Render()
	return tableString.String()
}

var deprecatedRuntimes = [7]string{
	"nodejs8.10", "nodejs6.10", "nodejs4.3", "nodejs4.3-edge", "nodejs", "dotnetcore2.0", "dotnetcore1.0",
}

func isDeprecatedRuntime(runtime string) bool {
	for _, r := range deprecatedRuntimes {
		if runtime == r {
			return true
		}
	}
	return false
}
