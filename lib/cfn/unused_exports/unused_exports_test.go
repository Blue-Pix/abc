package unused_exports_test

import (
	"errors"
	"testing"

	"github.com/Blue-Pix/abc/lib/cfn/unused_exports"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initMockClient(cm *unused_exports.MockCfnClient) {
	unused_exports.SetMockDefaultBehaviour(cm)
	unused_exports.CfnClient = cm
}

func TestFetchData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cmd := unused_exports.NewCmd()
		cm := &unused_exports.MockCfnClient{}
		initMockClient(cm)
		var args []string
		actual, err := unused_exports.FetchData(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		expected := []unused_exports.UnusedExport{
			{Name: "bar_key1", ExportingStack: "bar"},
			{Name: "foo_key2", ExportingStack: "foo"},
		}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
		cm.AssertNumberOfCalls(t, "ListStacks", 2)
		cm.AssertNumberOfCalls(t, "ListExports", 2)
		cm.AssertNumberOfCalls(t, "ListImports", 4)
	})

	t.Run("authorization error for ListStacks", func(t *testing.T) {
		cmd := unused_exports.NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListStacks operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListStacks"

		cm := &unused_exports.MockCfnClient{}
		cm.On("ListStacks", &cloudformation.ListStacksInput{NextToken: nil}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm)

		actual, err := unused_exports.FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStacks", 1)
		cm.AssertNumberOfCalls(t, "ListExports", 0)
		cm.AssertNumberOfCalls(t, "ListImports", 0)
	})

	t.Run("authorization error for ListExports", func(t *testing.T) {
		cmd := unused_exports.NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListExports operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListExports"

		cm := &unused_exports.MockCfnClient{}
		cm.On("ListExports", &cloudformation.ListExportsInput{NextToken: nil}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm)

		actual, err := unused_exports.FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStacks", 2)
		cm.AssertNumberOfCalls(t, "ListExports", 1)
		cm.AssertNumberOfCalls(t, "ListImports", 0)
	})

	t.Run("authorization error for ListImports", func(t *testing.T) {
		cmd := unused_exports.NewCmd()
		var args []string
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the ListImports operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:ListImports"

		cm := &unused_exports.MockCfnClient{}
		cm.On("ListImports", mock.AnythingOfType("*cloudformation.ListImportsInput")).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(cm)

		actual, err := unused_exports.FetchData(cmd, args)
		assert.Empty(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "ListStacks", 2)
		cm.AssertNumberOfCalls(t, "ListExports", 2)
		cm.AssertNumberOfCalls(t, "ListImports", 1)
	})
}
