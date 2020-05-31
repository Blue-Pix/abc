package purge_stack

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/spf13/cobra"
)

var CfnClient cloudformationiface.CloudFormationAPI
var EcrClient ecriface.ECRAPI

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge-stack",
		Short: "Delete stack completely including resource contents.",
		Long: `
[abc cfn purge-stack]
This command delete CloudFormation's stack,
which delete-stack api provided by AWS officcially cannot to perform.
For example, a stack which includes non-empty ECR repository.

Internally it uses aws cloudformation api.
Please configure your aws credentials with following policies.
- cloudformation:DeleteStack
- cloudformation:ListStackResources
- ecr:BatchDeleteImages
- ecr:DescribeImages`,
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
	str, err := toJSON(data)
	if err != nil {
		return err
	}
	cmd.Println(str)
	return nil
}

func FetchData(cmd *cobra.Command, args []string) ([]string, error) {
	initClient(cmd)

	params := &cloudformation.ListStackResourcesInput{
		StackName: aws.String("stack-with-ecr"),
	}
	resp, err := CfnClient.ListStackResources(params)
	if err != nil {
		return nil, err
	}
	for _, r := range resp.StackResourceSummaries {
		if aws.StringValue(r.ResourceType) == "AWS::ECR::Repository" {
			params2 := &ecr.DescribeImagesInput{
				MaxResults:     aws.Int64(1000),
				RepositoryName: r.PhysicalResourceId,
			}
			resp2, err := EcrClient.DescribeImages(params2)
			if err != nil {
				return nil, err
			}
			var images []*ecr.ImageIdentifier
			for _, i := range resp2.ImageDetails {
				images = append(images, &ecr.ImageIdentifier{
					ImageDigest: i.ImageDigest,
				})
			}
			params3 := &ecr.BatchDeleteImageInput{
				ImageIds:       images,
				RepositoryName: r.PhysicalResourceId,
			}
			if len(images) == 0 {
				continue
			}
			resp3, err := EcrClient.BatchDeleteImage(params3)
			if err != nil {
				return nil, err
			}
			if len(resp3.Failures) > 0 {
				cmd.Println(resp3.Failures)
				return nil, errors.New(fmt.Sprintf("failed to delete images of %s", aws.StringValue(r.PhysicalResourceId)))
			}
		}
	}

	return []string{}, nil
}

func initClient(cmd *cobra.Command) {
	profile, _ := cmd.Flags().GetString("profile")
	region, _ := cmd.Flags().GetString("region")
	if CfnClient == nil {
		sess := util.CreateSession(profile, region)
		CfnClient = cloudformation.New(sess)
	}
	if EcrClient == nil {
		sess := util.CreateSession(profile, region)
		EcrClient = ecr.New(sess)
	}
}

func toJSON(data []string) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)
	return jsonStr, nil
}
