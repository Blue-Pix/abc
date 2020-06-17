package create_stack

import (
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/mock"
)

type MockCfnClient struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (client *MockCfnClient) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	args := client.Called(input)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.CreateStackOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

type MockS3Client struct {
	mock.Mock
	s3iface.S3API
}

func (client *MockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	args := client.Called(input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func SetMockDefaultBehaviour(cm *MockCfnClient, sm *MockS3Client) {
	stackName := "foo"
	cm.On("CreateStack", &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
	}).Return(
		&cloudformation.CreateStackOutput{
			StackId: aws.String("xxxxxxxxxx"),
		},
		nil,
	)
	sm.On("GetObject", &s3.GetObjectInput{
		Bucket: aws.String("bucket_name"),
		Key:    aws.String("key"),
	}).Return(
		&s3.GetObjectOutput{
			Body: ioutil.NopCloser(strings.NewReader("hello world")),
		},
		nil,
	)
}