package purge_stack

import (
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

var (
	stackName string
)

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
	cmd.Flags().StringVar(&stackName, "stack-name", "", "stack name to delete")
	cmd.MarkFlagRequired("stack-name")
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if err := ExecPurgeStack(cmd, args); err != nil {
		return err
	}
	cmd.Println("Perform delete-stack is in progress asynchronously.\nPlease check deletion status by yourself.")
	return nil
}

func ExecPurgeStack(cmd *cobra.Command, args []string) error {
	initClient(cmd)
	resources, err := listEcrResources(nil, []*cloudformation.StackResourceSummary{})
	if err != nil {
		return err
	}
	for _, resource := range resources {
		repositoryName := resource.PhysicalResourceId
		images, err := listImageDigests(nil, repositoryName, []*ecr.ImageIdentifier{})
		if err != nil {
			return err
		}
		if len(images) == 0 {
			continue
		}
		failures, err := deleteImages(images, repositoryName)
		if err != nil {
			return err
		}
		if len(failures) > 0 {
			cmd.Println(failures)
			return errors.New(fmt.Sprintf("failed to delete images of %s", aws.StringValue(repositoryName)))
		}
		cmd.Println(fmt.Sprintf("All images in %s successfully deleted.", aws.StringValue(repositoryName)))
	}

	if err = deleteStack(stackName); err != nil {
		return err
	}

	return nil
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

func listEcrResources(token *string, ecrs []*cloudformation.StackResourceSummary) ([]*cloudformation.StackResourceSummary, error) {
	params := &cloudformation.ListStackResourcesInput{
		NextToken: token,
		StackName: aws.String(stackName),
	}
	resp, err := CfnClient.ListStackResources(params)
	if err != nil {
		return nil, err
	}
	for _, r := range resp.StackResourceSummaries {
		if aws.StringValue(r.ResourceType) == "AWS::ECR::Repository" {
			ecrs = append(ecrs, r)
		}
	}
	if resp.NextToken != nil {
		ecrs, err = listEcrResources(resp.NextToken, ecrs)
		if err != nil {
			return nil, err
		}
	}
	return ecrs, nil
}

func listImageDigests(token *string, repositoryName *string, images []*ecr.ImageIdentifier) ([]*ecr.ImageIdentifier, error) {
	params := &ecr.DescribeImagesInput{
		NextToken:      token,
		MaxResults:     aws.Int64(1000),
		RepositoryName: repositoryName,
	}
	resp, err := EcrClient.DescribeImages(params)
	if err != nil {
		return nil, err
	}
	for _, i := range resp.ImageDetails {
		images = append(images, &ecr.ImageIdentifier{
			ImageDigest: i.ImageDigest,
		})
	}
	if resp.NextToken != nil {
		images, err = listImageDigests(resp.NextToken, repositoryName, images)
		if err != nil {
			return nil, err
		}
	}
	return images, nil
}

func deleteImages(images []*ecr.ImageIdentifier, repositoryName *string) ([]*ecr.ImageFailure, error) {
	params := &ecr.BatchDeleteImageInput{
		ImageIds:       images,
		RepositoryName: repositoryName,
	}
	resp, err := EcrClient.BatchDeleteImage(params)
	if err != nil {
		return nil, err
	}
	return resp.Failures, nil
}

func deleteStack(stackName string) error {
	params := &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}
	_, err := CfnClient.DeleteStack(params)
	if err != nil {
		return err
	}
	return nil
}
