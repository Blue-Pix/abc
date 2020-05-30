package unused_exports

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	cloudformationiface.CloudFormationAPI
	ListStacksMock  func(params *cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error)
	ListExportsMock func(params *cloudformation.ListExportsInput) (*cloudformation.ListExportsOutput, error)
	ListImportsMock func(params *cloudformation.ListImportsInput) (*cloudformation.ListImportsOutput, error)
}

func (client *mockClient) ListStacks(params *cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error) {
	if client.ListStacksMock != nil {
		return client.ListStacksMock(params)
	}

	switch aws.StringValue(params.NextToken) {
	case "":
		return &cloudformation.ListStacksOutput{
			NextToken: aws.String("next_token"),
			StackSummaries: []*cloudformation.StackSummary{
				{StackId: aws.String("aaa"), StackName: aws.String("foo")},
				{StackId: aws.String("bbb"), StackName: aws.String("bar")},
			},
		}, nil

	case "next_token":
		return &cloudformation.ListStacksOutput{
			NextToken: nil,
			StackSummaries: []*cloudformation.StackSummary{
				{StackId: aws.String("ccc"), StackName: aws.String("foobar")},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (client *mockClient) ListExports(params *cloudformation.ListExportsInput) (*cloudformation.ListExportsOutput, error) {
	if client.ListExportsMock != nil {
		return client.ListExportsMock(params)
	}

	switch aws.StringValue(params.NextToken) {
	case "":
		return &cloudformation.ListExportsOutput{
			NextToken: aws.String("next_token"),
			Exports: []*cloudformation.Export{
				{Name: aws.String("foo_key1"), ExportingStackId: aws.String("aaa")},
				{Name: aws.String("foo_key2"), ExportingStackId: aws.String("aaa")},
			},
		}, nil

	case "next_token":
		return &cloudformation.ListExportsOutput{
			NextToken: nil,
			Exports: []*cloudformation.Export{
				{Name: aws.String("bar_key1"), ExportingStackId: aws.String("bbb")},
				{Name: aws.String("bar_key2"), ExportingStackId: aws.String("bbb")},
			},
		}, nil
	default:
		return nil, nil
	}
}

func (client *mockClient) ListImports(params *cloudformation.ListImportsInput) (*cloudformation.ListImportsOutput, error) {
	if client.ListImportsMock != nil {
		return client.ListImportsMock(params)
	}

	switch aws.StringValue(params.ExportName) {
	case "foo_key1":
		return &cloudformation.ListImportsOutput{
			NextToken: nil,
			Imports: []*string{
				aws.String("bar"),
				aws.String("foobar"),
			},
		}, nil
	case "bar_key2":
		return &cloudformation.ListImportsOutput{
			NextToken: nil,
			Imports: []*string{
				aws.String("foobar"),
			},
		}, nil
	default:
		return nil, awserr.New(
			"ValidationError",
			fmt.Sprintf("%s is not imported by any stack", aws.StringValue(params.ExportName)),
			errors.New("hoge"),
		)
	}
}

func TestMain(m *testing.M) {
	Client = &mockClient{}
	code := m.Run()
	os.Exit(code)
}

func TestFetchData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cmd := NewCmd()
		var args []string
		actual, err := FetchData(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		expected := []UnusedExport{
			{Name: "bar_key1", ExportingStack: "bar"},
			{Name: "foo_key2", ExportingStack: "foo"},
		}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})

	t.Run("authorization error for ListStacks", func(t *testing.T) {
		cmd := NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListStacks operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListStacks"
		Client = &mockClient{
			ListStacksMock: func(params *cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error) {
				return nil, awserr.New(errorCode, errorMsg, errors.New("hoge"))
			},
		}
		actual, err := FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
	})

	t.Run("authorization error for ListExports", func(t *testing.T) {
		cmd := NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListExports operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListExports"
		Client = &mockClient{
			ListExportsMock: func(params *cloudformation.ListExportsInput) (*cloudformation.ListExportsOutput, error) {
				return nil, awserr.New(errorCode, errorMsg, errors.New("hoge"))
			},
		}
		actual, err := FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
	})

	t.Run("authorization error for ListImports", func(t *testing.T) {
		cmd := NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListImports operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListImports"
		Client = &mockClient{
			ListImportsMock: func(params *cloudformation.ListImportsInput) (*cloudformation.ListImportsOutput, error) {
				return nil, awserr.New(errorCode, errorMsg, errors.New("hoge"))
			},
		}
		actual, err := FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
	})
}
