package unusedExports

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unused-exports",
		Short: "list all exports which not used in any stack.",
		Long: `
		[abc cfn unused-exports]
		This command returns all CloudFormation's exports name,
		which not used in any stack.
			
		Internally it uses aws cloudformation api.
		Please configure your aws credentials with following policies.
		- cloudformation:ListExports
		- cloudformation:ListImports
		- cloudformation:ListStacks

		By default, it prints exports names separated by comma.
		You can customize delimiter with -d option.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	str, err := Run(cmd, args)
	if err != nil {
		return err
	}
	cmd.Println(str)
	return nil
}

func Run(cmd *cobra.Command, args []string) (string, error) {
	sess := session.Must(session.NewSession())
	// sess := session.Must(
	// 	session.NewSessionWithOptions(
	// 		session.Options{
	// 			Profile: "profile",
	// 		},
	// 	),
	// )

	stacks := make(map[string]string)
	if err := listStacks(sess, nil, stacks); err != nil {
		return "", err
	}
	exports := make(map[string]string)
	if err := listExports(sess, nil, exports); err != nil {
		return "", err
	}

	var csv []string
	csv = append(csv, "name,exporting_stack")
	for key, _ := range exports {
		var result []string
		if err := listImports(sess, key, nil, &result); err != nil {
			return "", err
		}
		if len(result) == 0 {
			csv = append(csv, fmt.Sprintf("%s,%s", key, stacks[exports[key]]))
		}
	}
	return strings.Join(csv, "\n"), nil
}

func listStacks(sess *session.Session, token *string, result map[string]string) error {
	time.Sleep(1 * time.Second)
	service := cloudformation.New(sess)
	params := &cloudformation.ListStacksInput{
		NextToken: token,
	}
	resp, err := service.ListStacks(params)
	if err != nil {
		return err
	}
	for _, stack := range resp.StackSummaries {
		result[aws.StringValue(stack.StackId)] = aws.StringValue(stack.StackName)
	}
	if resp.NextToken != nil {
		if err = listStacks(sess, resp.NextToken, result); err != nil {
			return err
		}
	}
	return nil
}

func listExports(sess *session.Session, token *string, result map[string]string) error {
	time.Sleep(1 * time.Second)
	service := cloudformation.New(sess)
	params := &cloudformation.ListExportsInput{
		NextToken: token,
	}
	resp, err := service.ListExports(params)
	if err != nil {
		return err
	}
	for _, export := range resp.Exports {
		result[aws.StringValue(export.Name)] = aws.StringValue(export.ExportingStackId)
	}
	if resp.NextToken != nil {
		if err = listExports(sess, resp.NextToken, result); err != nil {
			return err
		}
	}
	return nil
}

func listImports(sess *session.Session, exportName string, token *string, result *[]string) error {
	time.Sleep(1 * time.Second)
	service := cloudformation.New(sess)
	params := &cloudformation.ListImportsInput{
		NextToken:  token,
		ExportName: aws.String(exportName),
	}
	resp, err := service.ListImports(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == "ValidationError" && strings.Contains(aerr.Message(), "is not imported by any stack") {
				return nil
			}
		}
		return err
	}
	for _, _import := range resp.Imports {
		*result = append(*result, aws.StringValue(_import))
	}
	if resp.NextToken != nil {
		if err = listImports(sess, exportName, resp.NextToken, result); err != nil {
			return err
		}
	}
	return nil
}
