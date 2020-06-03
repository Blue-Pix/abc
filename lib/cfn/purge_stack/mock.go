package purge_stack

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/stretchr/testify/mock"
)

type MockCfnClient struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (client *MockCfnClient) ListStackResources(params *cloudformation.ListStackResourcesInput) (*cloudformation.ListStackResourcesOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.ListStackResourcesOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *MockCfnClient) DeleteStack(params *cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.DeleteStackOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

type MockEcrClient struct {
	mock.Mock
	ecriface.ECRAPI
}

func (client *MockEcrClient) DescribeImages(params *ecr.DescribeImagesInput) (*ecr.DescribeImagesOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *MockEcrClient) BatchDeleteImage(params *ecr.BatchDeleteImageInput) (*ecr.BatchDeleteImageOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.BatchDeleteImageOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func SetMockDefaultBehaviour(cm *MockCfnClient, em *MockEcrClient) {
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
		NextToken:      nil,
		MaxResults:     aws.Int64(1000),
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
		NextToken:      aws.String("next_token"),
		MaxResults:     aws.Int64(1000),
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
		NextToken:      nil,
		MaxResults:     aws.Int64(1000),
		RepositoryName: aws.String("ecr2"),
	}).Return(
		&ecr.DescribeImagesOutput{
			NextToken:    nil,
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
}
