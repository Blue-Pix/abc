package unused_exports

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/mock"
)

type MockCfnClient struct {
	mock.Mock
	cloudformationiface.CloudFormationAPI
}

func (client *MockCfnClient) ListStacks(params *cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.ListStacksOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *MockCfnClient) ListExports(params *cloudformation.ListExportsInput) (*cloudformation.ListExportsOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.ListExportsOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (client *MockCfnClient) ListImports(params *cloudformation.ListImportsInput) (*cloudformation.ListImportsOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudformation.ListImportsOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func SetMockDefaultBehaviour(cm *MockCfnClient) {
	cm.On("ListStacks", &cloudformation.ListStacksInput{
		NextToken: nil,
	}).Return(
		&cloudformation.ListStacksOutput{
			NextToken: aws.String("next_token"),
			StackSummaries: []*cloudformation.StackSummary{
				{StackId: aws.String("aaa"), StackName: aws.String("foo")},
				{StackId: aws.String("bbb"), StackName: aws.String("bar")},
			},
		},
		nil,
	)
	cm.On("ListStacks", &cloudformation.ListStacksInput{
		NextToken: aws.String("next_token"),
	}).Return(
		&cloudformation.ListStacksOutput{
			NextToken: nil,
			StackSummaries: []*cloudformation.StackSummary{
				{StackId: aws.String("ccc"), StackName: aws.String("foobar")},
			},
		},
		nil,
	)
	cm.On("ListExports", &cloudformation.ListExportsInput{
		NextToken: nil,
	}).Return(
		&cloudformation.ListExportsOutput{
			NextToken: aws.String("next_token"),
			Exports: []*cloudformation.Export{
				{Name: aws.String("foo_key1"), ExportingStackId: aws.String("aaa")},
				{Name: aws.String("foo_key2"), ExportingStackId: aws.String("aaa")},
			},
		},
		nil,
	)
	cm.On("ListExports", &cloudformation.ListExportsInput{
		NextToken: aws.String("next_token"),
	}).Return(
		&cloudformation.ListExportsOutput{
			NextToken: nil,
			Exports: []*cloudformation.Export{
				{Name: aws.String("bar_key1"), ExportingStackId: aws.String("bbb")},
				{Name: aws.String("bar_key2"), ExportingStackId: aws.String("bbb")},
			},
		},
		nil,
	)
	cm.On("ListImports", &cloudformation.ListImportsInput{
		ExportName: aws.String("foo_key1"),
	}).Return(
		&cloudformation.ListImportsOutput{
			NextToken: nil,
			Imports: []*string{
				aws.String("bar"),
				aws.String("foobar"),
			},
		},
		nil,
	)
	cm.On("ListImports", &cloudformation.ListImportsInput{
		ExportName: aws.String("bar_key2"),
	}).Return(
		&cloudformation.ListImportsOutput{
			NextToken: nil,
			Imports: []*string{
				aws.String("foobar"),
			},
		},
		nil,
	)
	cm.On("ListImports", &cloudformation.ListImportsInput{
		ExportName: aws.String("foo_key2"),
	}).Return(
		nil,
		awserr.New(
			"ValidationError",
			"foo_key2 is not imported by any stack",
			errors.New("hoge"),
		),
	)
	cm.On("ListImports", &cloudformation.ListImportsInput{
		ExportName: aws.String("bar_key1"),
	}).Return(
		nil,
		awserr.New(
			"ValidationError",
			"bar_key1 is not imported by any stack",
			errors.New("hoge"),
		),
	)
}
