package stats

import (
	"encoding/json"
	"errors"
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

var (
	verbose bool
	format  string
)

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
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detail")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "output format (table or json)")
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	data, err := FetchData(cmd, args)
	if err != nil {
		return err
	}
	var str string
	if len(data) == 0 && format == "table" {
		str = "no function found."
	} else {
		str, err = Output(data)
		if err != nil {
			return err
		}
	}
	cmd.Println(str)
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

func Output(count map[string][]string) (string, error) {
	if format == "json" {
		return jsonOutput(count)
	} else if format == "table" {
		return tableOutput(count), nil
	}
	return "", errors.New("invalid format.")
}

func jsonOutput(count map[string][]string) (string, error) {
	if verbose {
		return verboseJsonOutput(count)
	} else {
		return normalJsonOutput(count)
	}
}
func normalJsonOutput(count map[string][]string) (string, error) {
	keys := sortKey(count)

	type Stats struct {
		Runtime    string `json:"runtime"`
		Count      int    `json:"count"`
		Deprecated bool   `json:"deprecated"`
	}

	statsList := make([]Stats, len(keys))
	for i, k := range keys {
		statsList[i] = Stats{
			Runtime:    k,
			Count:      len(count[k]),
			Deprecated: isDeprecatedRuntime(k),
		}
	}
	jsonBytes, err := json.Marshal(statsList)
	return string(jsonBytes), err
}
func verboseJsonOutput(count map[string][]string) (string, error) {
	keys := sortKey(count)

	type VerboseStats struct {
		Runtime    string   `json:"runtime"`
		Count      int      `json:"count"`
		Functions  []string `json:"functions"`
		Deprecated bool     `json:"deprecated"`
	}

	statsList := make([]VerboseStats, len(keys))
	for i, k := range keys {
		statsList[i] = VerboseStats{
			Runtime:    k,
			Count:      len(count[k]),
			Functions:  count[k],
			Deprecated: isDeprecatedRuntime(k),
		}
	}
	jsonBytes, err := json.Marshal(statsList)
	return string(jsonBytes), err
}

func tableOutput(count map[string][]string) string {
	keys := sortKey(count)
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	if verbose {
		verboseTableOutput(keys, table, count)
	} else {
		normalTableOutput(keys, table, count)
	}
	return tableString.String()
}

func normalTableOutput(keys []string, table *tablewriter.Table, count map[string][]string) {
	table.SetHeader([]string{"Runtime", "Count"})
	for _, k := range keys {
		runtime := k
		if isDeprecatedRuntime(runtime) {
			runtime = fmt.Sprintf("\x1b[91m%s（Deprecated）\x1b[0m", runtime)
		}
		table.Append([]string{runtime, strconv.Itoa(len(count[k]))})
	}
	table.Render()
}

func verboseTableOutput(keys []string, table *tablewriter.Table, count map[string][]string) {
	table.SetHeader([]string{"Runtime", "Count", "Functions"})
	for _, k := range keys {
		runtime := k
		if isDeprecatedRuntime(runtime) {
			runtime = fmt.Sprintf("\x1b[91m%s（Deprecated）\x1b[0m", runtime)
		}
		table.Append([]string{runtime, strconv.Itoa(len(count[k])), strings.Join(count[k], ", ")})
	}
	table.Render()
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
