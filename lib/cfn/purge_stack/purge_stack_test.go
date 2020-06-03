package purge_stack

import (
	"fmt"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCfnClient struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (client *mockCfnClient) ListStackResources(params *cloudformation.ListStackResourcesInput) (*cloudformation.ListStackResourcesOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.ListStackResourcesOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *mockCfnClient) DeleteStack(params *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.DeleteStackOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}
type mockEcrClient struct {
	mock.Mock
	ecriface.ECRAPI
}

func (client *mockEcrClient) DescribeImages(params *ecr.DescribeImagesInput) (*ecr.DescribeImagesOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *mockEcrClient) BatchDeleteImage(params *ecr.BatchDeleteImageInput) (*ecr.BatchDeleteImageOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.BatchDeleteImageOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func initMockClient(cm *mockCfnClient, em *mockEcrClient) {
	stackName := "foo"
	cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{
		StackName: aws.String(stackName),
	}).Return(
		&cloudformation.ListStackResourcesOutput{
			NextToken: aws.String("next_token"),
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{PhysicalResourceId: aws.String("cluster"), ResourceType: aws.String("AWS::ECS::Cluster")},
				{PhysicalResourceId: aws.String("ecr1"), ResourceType: aws.String("AWS::ECR::Repository")},
			},
		}, 
		nil,
	)
	cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{
		NextToken: aws.String("next_token"), 
		StackName: aws.String(stackName),
	}).Return(
		&cloudformation.ListStackResourcesOutput{
			NextToken: nil,
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{PhysicalResourceId: aws.String("ecr2"), ResourceType: aws.String("AWS::ECR::Repository")},
				{PhysicalResourceId: aws.String("queue"), ResourceType: aws.String("AWS::SQS::Queue")},
			},
		}, 
		nil,
	)
	em.On("DescribeImages", &ecr.DescribeImagesInput{
		NextToken: nil, 
		MaxResults: aws.Int64(1000), 
		RepositoryName: aws.String("ecr1"),
	}).Return(
		&ecr.DescribeImagesOutput{
			NextToken: aws.String("next_token"),
			ImageDetails: []*ecr.ImageDetail{
				{ImageDigest: aws.String("foofoofoo"), ImageTags: []*string{aws.String("foo")}},
				{ImageDigest: aws.String("barbarbar"), ImageTags: []*string{aws.String("bar")}},
			},
		}, 
		nil,
	)
	em.On("DescribeImages", &ecr.DescribeImagesInput{
		NextToken: aws.String("next_token"), 
		MaxResults: aws.Int64(1000), 
		RepositoryName: aws.String("ecr1"),
	}).Return(
		&ecr.DescribeImagesOutput{
			NextToken: nil,
			ImageDetails: []*ecr.ImageDetail{
				{ImageDigest: aws.String("foobarfoobar"), ImageTags: []*string{aws.String("foobar")}},
				{ImageDigest: aws.String("barfoobarfoo"), ImageTags: []*string{aws.String("barfoo")}},
			},
		}, 
		nil,
	)
	em.On("DescribeImages", &ecr.DescribeImagesInput{
		NextToken: nil, 
		MaxResults: aws.Int64(1000), 
		RepositoryName: aws.String("ecr2"),
	}).Return(
		&ecr.DescribeImagesOutput{
			NextToken: nil, 
			ImageDetails: []*ecr.ImageDetail{},
		},
		nil,
	)
	em.On("BatchDeleteImage", &ecr.BatchDeleteImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{ImageDigest: aws.String("foofoofoo")},
			{ImageDigest: aws.String("barbarbar")},
			{ImageDigest: aws.String("foobarfoobar")},
			{ImageDigest: aws.String("barfoobarfoo")},
		},
		RepositoryName: aws.String("ecr1"),
	}).Return(
		&ecr.BatchDeleteImageOutput{
			ImageIds: []*ecr.ImageIdentifier{},
			Failures: []*ecr.ImageFailure{},
		}, 
		nil,
	)
	cm.On("DeleteStack", &cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	}).Return(
		&cloudformation.DeleteStackOutput{}, 
		nil,
	)
	CfnClient = cm
	EcrClient = em
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestExecPurgeStack(t *testing.T) {
	// include two ecr resources, one with images, other with no image.
	t.Run("success", func(t *testing.T) {
		stackName := "foo"
		cm := &mockCfnClient{}
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

		assert.Nil(t, err)
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 3)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 1)
	})

	t.Run("stack without ecr", func(t *testing.T) {
		stackName := "foo"
		cm := &mockCfnClient{}
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
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		em := &mockEcrClient{}
		em.On("DescribeImages", &ecr.DescribeImagesInput{NextToken: nil, MaxResults: aws.Int64(1000), RepositoryName: aws.String("ecr1")}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		em := &mockEcrClient{}
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

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		cm.On("DeleteStack", &cloudformation.DeleteStackInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		em := &mockEcrClient{}
		initMockClient(cm, em)

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

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
		cm := &mockCfnClient{}
		em := &mockEcrClient{}
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

		cmd := NewCmd()
		cmd.Flags().Set("stack-name", stackName)
		var args []string
		err := ExecPurgeStack(cmd, args)

		assert.Equal(t, fmt.Sprintf("failed to delete images of %s", "ecr1"), err.Error())
		cm.AssertNumberOfCalls(t, "ListStackResources", 2)
		em.AssertNumberOfCalls(t, "DescribeImages", 2)
		em.AssertNumberOfCalls(t, "BatchDeleteImage", 1)
		cm.AssertNumberOfCalls(t, "DeleteStack", 0)
	})
}
