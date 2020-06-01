package purge_stack

import (
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
	_ = client.Called(params)

	switch aws.StringValue(params.NextToken) {
	case "":
		return &cloudformation.ListStackResourcesOutput{
			NextToken: aws.String("next_token"),
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{PhysicalResourceId: aws.String("cluster"), ResourceType: aws.String("AWS::ECS::Cluster")},
				{PhysicalResourceId: aws.String("ecr1"), ResourceType: aws.String("AWS::ECR::Repository")},
			},
		}, nil

	case "next_token":
		return &cloudformation.ListStackResourcesOutput{
			NextToken: nil,
			StackResourceSummaries: []*cloudformation.StackResourceSummary{
				{PhysicalResourceId: aws.String("ecr2"), ResourceType: aws.String("AWS::ECR::Repository")},
				{PhysicalResourceId: aws.String("queue"), ResourceType: aws.String("AWS::SQS::Queue")},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (client *mockCfnClient) DeleteStack(params *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	args := client.Called(params)
	if len(args) > 1 && args.Get(1) != nil {
		return nil, args.Get(1).(awserr.Error)
	}
	return &cloudformation.DeleteStackOutput{}, nil
}

type mockEcrClient struct {
	mock.Mock
	ecriface.ECRAPI
}

func (client *mockEcrClient) DescribeImages(params *ecr.DescribeImagesInput) (*ecr.DescribeImagesOutput, error) {
	_ = client.Called(params)
	if aws.StringValue(params.RepositoryName) == "ecr2" {
		return &ecr.DescribeImagesOutput{NextToken: nil, ImageDetails: []*ecr.ImageDetail{}}, nil
	}

	switch aws.StringValue(params.NextToken) {
	case "":
		return &ecr.DescribeImagesOutput{
			NextToken: aws.String("next_token"),
			ImageDetails: []*ecr.ImageDetail{
				{ImageDigest: aws.String("foofoofoo"), ImageTags: []*string{aws.String("foo")}},
				{ImageDigest: aws.String("barbarbar"), ImageTags: []*string{aws.String("bar")}},
			},
		}, nil

	case "next_token":
		return &ecr.DescribeImagesOutput{
			NextToken: nil,
			ImageDetails: []*ecr.ImageDetail{
				{ImageDigest: aws.String("foobarfoobar"), ImageTags: []*string{aws.String("foobar")}},
				{ImageDigest: aws.String("barfoobarfoo"), ImageTags: []*string{aws.String("barfoo")}},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (client *mockEcrClient) BatchDeleteImage(params *ecr.BatchDeleteImageInput) (*ecr.BatchDeleteImageOutput, error) {
	_ = client.Called(params)
	return &ecr.BatchDeleteImageOutput{
		ImageIds: []*ecr.ImageIdentifier{},
		Failures: []*ecr.ImageFailure{},
	}, nil
}

func initMockClient(cm *mockCfnClient, em *mockEcrClient) {
	stackName := "foo"
	cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{StackName: aws.String(stackName)})
	cm.On("ListStackResources", &cloudformation.ListStackResourcesInput{NextToken: aws.String("next_token"), StackName: aws.String(stackName)})
	em.On("DescribeImages", &ecr.DescribeImagesInput{NextToken: nil, MaxResults: aws.Int64(1000), RepositoryName: aws.String("ecr1")})
	em.On("DescribeImages", &ecr.DescribeImagesInput{NextToken: aws.String("next_token"), MaxResults: aws.Int64(1000), RepositoryName: aws.String("ecr1")})
	em.On("DescribeImages", &ecr.DescribeImagesInput{NextToken: nil, MaxResults: aws.Int64(1000), RepositoryName: aws.String("ecr2")})
	em.On("BatchDeleteImage", &ecr.BatchDeleteImageInput{
		ImageIds: []*ecr.ImageIdentifier{
			{ImageDigest: aws.String("foofoofoo")},
			{ImageDigest: aws.String("barbarbar")},
			{ImageDigest: aws.String("foobarfoobar")},
			{ImageDigest: aws.String("barfoobarfoo")},
		},
		RepositoryName: aws.String("ecr1"),
	})
	cm.On("DeleteStack", &cloudformation.DeleteStackInput{StackName: aws.String(stackName)})
	CfnClient = cm
	EcrClient = em
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestExecPurgeStack(t *testing.T) {
	stackName := "foo"

	t.Run("success", func(t *testing.T) {
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

	t.Run("authorization error for DeleteStack", func(t *testing.T) {
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

}
