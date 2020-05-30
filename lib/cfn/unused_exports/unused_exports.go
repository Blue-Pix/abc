package unused_exports

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/spf13/cobra"
)

var Client cloudformationiface.CloudFormationAPI

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unused-exports",
		Short: "List all exports which not used in any stack.",
		Long: `
[abc cfn unused-exports]
This command returns all CloudFormation's exports name,
which not used in any stack, in csv format.
	
Internally it uses aws cloudformation api.
Please configure your aws credentials with following policies.
- cloudformation:ListExports
- cloudformation:ListImports
- cloudformation:ListStacks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	unused_exports, err := FetchData(cmd, args)
	if err != nil {
		return err
	}
	str, err := toJSON(unused_exports)
	if err != nil {
		return err
	}
	cmd.Println(str)
	return nil
}

func FetchData(cmd *cobra.Command, args []string) ([]UnusedExport, error) {
	initClient(cmd)

	stacks := make(map[string]string)
	if err := listStacks(nil, stacks); err != nil {
		return nil, err
	}
	exports := make(map[string]string)
	if err := listExports(nil, exports); err != nil {
		return nil, err
	}
	names, err := selectUnusedExportName(exports)
	if err != nil {
		return nil, err
	}
	var unused_exports []UnusedExport
	for _, name := range names {
		unused_exports = append(unused_exports, UnusedExport{Name: name, ExportingStack: stacks[exports[name]]})
	}
	return unused_exports, nil
}

func initClient(cmd *cobra.Command) {
	if Client == nil {
		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		sess := util.CreateSession(profile, region)
		Client = cloudformation.New(sess)
	}
}

type UnusedExport struct {
	Name           string `json:"name"`
	ExportingStack string `json:"exporting_stack"`
}

func listStacks(token *string, result map[string]string) error {
	time.Sleep(1 * time.Second)
	params := &cloudformation.ListStacksInput{
		NextToken: token,
	}
	resp, err := Client.ListStacks(params)
	if err != nil {
		return err
	}
	for _, stack := range resp.StackSummaries {
		result[aws.StringValue(stack.StackId)] = aws.StringValue(stack.StackName)
	}
	if resp.NextToken != nil {
		if err = listStacks(resp.NextToken, result); err != nil {
			return err
		}
	}
	return nil
}

func listExports(token *string, result map[string]string) error {
	time.Sleep(1 * time.Second)
	params := &cloudformation.ListExportsInput{
		NextToken: token,
	}
	resp, err := Client.ListExports(params)
	if err != nil {
		return err
	}
	for _, export := range resp.Exports {
		result[aws.StringValue(export.Name)] = aws.StringValue(export.ExportingStackId)
	}
	if resp.NextToken != nil {
		if err = listExports(resp.NextToken, result); err != nil {
			return err
		}
	}
	return nil
}

func listImports(exportName string, token *string, result []string) ([]string, error) {
	time.Sleep(1 * time.Second)
	params := &cloudformation.ListImportsInput{
		NextToken:  token,
		ExportName: aws.String(exportName),
	}
	resp, err := Client.ListImports(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "ValidationError" && strings.Contains(aerr.Message(), "is not imported by any stack") {
				return nil, nil
			}
		}
		return nil, err
	}
	var new_result []string
	for _, _import := range resp.Imports {
		new_result = append(result, aws.StringValue(_import))
	}
	if resp.NextToken != nil {
		new_result2, err := listImports(exportName, resp.NextToken, new_result)
		if err != nil {
			return nil, err
		}
		new_result = append(new_result, new_result2...)
	}
	return new_result, nil
}

func selectUnusedExportName(exports map[string]string) ([]string, error) {
	var keys []string
	for key := range exports {
		result, err := listImports(key, nil, []string{})
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys, nil
}

func toJSON(exports []UnusedExport) (string, error) {
	jsonBytes, err := json.Marshal(exports)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)
	return jsonStr, nil
}
