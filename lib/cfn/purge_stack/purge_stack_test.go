package purge_stack_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Blue-Pix/abc/lib/cfn/purge_stack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initMockClient(cm *purge_stack.MockCfnClient, em *purge_stack.MockEcrClient) {
	purge_stack.SetMockDefaultBehaviour(cm, em)
	purge_stack.CfnClient = cm
	purge_stack.EcrClient = em
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestExecPurgeStack(t *testing.T) {
	// include two ecr resources, one with images, other with no image.
	t.Run("success", func(t *testing.T) {
		stackName := "foo"
		cm := &purge_stack.MockCfnClient{}
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Nil(t, err)
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 3)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 1)
	})

	t.Run("stack without ecr", func(t *testing.T) {
		stackName := "foo"
		cm := &purge_stack.MockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{
			StackName: aws.String(stackName),
		}).Return(
			&cloudformation.ListStackResourcesOutput{
				NextToken: nil,
				StackResourceSummaries: []*cloudformation.StackResourceSummary{
					{PhysicalResourceId: aws.String("cluster"), ResourceType: aws.String("AWS::ECS::Cluster")},
				},
			},
			nil,
		)
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Nil(t, err)
		cm.AssertNumberOfCalls(t, "ListStackResources", 1)
		em.AssertNumberOfCalls(t, "DescribeImages", 0)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 0)
		cm.AssertNumberOfCalls(t, "DeleteStack", 1)
	})

	/************************************
		Authorization Error
	************************************/

	t.Run("authorization error for ListStackResources", func(t *testing.T) {
		stackName := "foo"
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListStackResources operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListStackResources"
		cm := &purge_stack.MockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 1)
		em.AssertNumberOfCalls(t, "DescribeImages", 0)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 0)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})

	t.Run("authorization error for DescribeImages", func(t *testing.T) {
		stackName := "foo"
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the DescribeImages operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: ecr:DescribeImages"
		cm := &purge_stack.MockCfnClient{}
		em := &purge_stack.MockEcrClient{}
		em.On("DescribeImages", &ecr.DescribeImagesInput{NextToken: nil, MaxResults: aws.Int64(1000), RepositoryName: aws.String("ecr1")}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 1)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 0)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})

	t.Run("authorization error for BatchDeleteImage", func(t *testing.T) {
		stackName := "foo"
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the BatchDeleteImage operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: ecr:BatchDeleteImage"
		cm := &purge_stack.MockCfnClient{}
		em := &purge_stack.MockEcrClient{}
		em.On("BatchDeleteImage", &ecr.BatchDeleteImageInput{
			ImageIds: []*ecr.ImageIdentifier{
				{ImageDigest: aws.String("foofoofoo")},
				{ImageDigest: aws.String("barbarbar")},
				{ImageDigest: aws.String("foobarfoobar")},
				{ImageDigest: aws.String("barfoobarfoo")},
			},
			RepositoryName: aws.String("ecr1"),
		}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 2)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})

	t.Run("authorization error for DeleteStack", func(t *testing.T) {
		stackName := "foo"
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListStacks operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListStacks"
		cm := &purge_stack.MockCfnClient{}
		cm.On("DeleteStack", &cloudformation.DeleteStackInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 3)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 1)
	})

	t.Run("authorization error for ListStackResources", func(t *testing.T) {
		stackName := "foo"
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListStackResources operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListStackResources"
		cm := &purge_stack.MockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 1)
		em.AssertNumberOfCalls(t, "DescribeImages", 0)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 0)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})

	/************************************
		Validation Error
	************************************/

	t.Run("no such stack name", func(t *testing.T) {
		stackName := "no_such_stack_name"
		const errorCode = "ValidationError"
		const errorMsg = "Stack with id no_such_stack_name does not exist"
		cm := &purge_stack.MockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &purge_stack.MockEcrClient{}
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStackResources", 1)
		em.AssertNumberOfCalls(t, "DescribeImages", 0)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 0)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})

	/************************************
		Others
	************************************/

	t.Run("failed on BatchDeleteImage", func(t *testing.T) {
		stackName := "foo"
		cm := &purge_stack.MockCfnClient{}
		em := &purge_stack.MockEcrClient{}
		em.On("BatchDeleteImage", mock.AnythingOfType("*ecr.BatchDeleteImageInput")).Return(
			&ecr.BatchDeleteImageOutput{
				ImageIds: []*ecr.ImageIdentifier{},
				Failures: []*ecr.ImageFailure{
					{FailureCode: aws.String("hoge"), FailureReason: aws.String("fuga"), ImageId: &ecr.ImageIdentifier{}},
				},
			},
			nil,
		)
		initMockClient(cm, em)

		cmd := purge_stack.NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := purge_stack.ExecPurgeStack(cmd, args)

		assert.Equal(t, fmt.Sprintf("failed to delete images of %s", "ecr1"), err.Error())
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 2)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})
}
