package stats_test

import (
	"errors"
	"testing"

	"github.com/Blue-Pix/abc/lib/lambda/stats"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/assert"
)

func initMockClient(lm *stats.MockLambdaClient) {
	stats.SetMockDefaultBehaviour(lm)
	stats.LambdaClient = lm
}

func TestFetchData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lm := &stats.MockLambdaClient{}
		initMockClient(lm)

		expected := map[string][]string{
			"nodejs12.x": {
				"node12-func-1",
			},
			"nodejs10.x": {
				"node10-func-1",
			},
			"python3.8": {
				"python3.8-func-1",
				"python3.8-func-2",
				"python3.8-func-3",
			},
			"python3.7": {
				"python3.7-func-1",
			},
			"python3.6": {
				"python3.6-func-1",
			},
			"ruby2.7": {
				"ruby2.7-func-1",
			},
			"ruby2.5": {
				"ruby2.5-func-1",
			},
			"java11": {
				"java11-func-1",
			},
			"java8": {
				"java8-func-1",
			},
			"go1.x": {
				"go1-func-1",
				"go1-func-2",
				"go1-func-3",
				"go1-func-4",
			},
			"dotnetcore3.1": {
				"dotnet3.1-func-1",
			},
			"dotnetcore2.1": {
				"dotnet2.1-func-1",
			},
			"provided": {
				"provided-func-1",
				"provided-func-2",
				"provided-func-3",
				"provided-func-4",
				"provided-func-5",
				"provided-func-6",
			},
			"nodejs4.3": {
				"nodejs4.3-func-1",
			},
			"nodejs4.3-edge": {
				"nodejs4.3edge-func-1",
			},
			"nodejs6.10": {
				"nodejs6.10-func-1",
			},
			"nodejs8.10": {
				"nodejs8.10-func-1",
			},
			"nodejs": {
				"nodejs0.10-func-1",
				"nodejs0.10-func-2",
			},
			"dotnetcore1.0": {
				"dotnet1.0-func-1",
			},
			"dotnetcore2.0": {
				"dotnet2.0-func-1",
			},
		}
		cmd := stats.NewCmd()
		var args []string
		actual, err := stats.FetchData(cmd, args)
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
		lm.AssertNumberOfCalls(t, "ListFunctions", 2)
	})

	/************************************
		Authorization Error
	************************************/

	t.Run("authorization error for ListFunctions", func(t *testing.T) {
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (ListFunctions) when calling the ListFunctions operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: lambda:ListFunctions"
		lm := &stats.MockLambdaClient{}
		lm.On("ListFunctions", &lambda.ListFunctionsInput{Marker: nil, MaxItems: aws.Int64(1000)}).Return(nil, awserr.New(errorCode, errorMsg, errors.New("hoge")))
		initMockClient(lm)

		cmd := stats.NewCmd()
		var args []string
		actual, err := stats.FetchData(cmd, args)
		assert.Nil(t, actual)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		lm.AssertNumberOfCalls(t, "ListFunctions", 1)
	})
}

func TestOutput(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lm := &stats.MockLambdaClient{}
		initMockClient(lm)

		expected := "|           RUNTIME            | COUNT |\n"
		expected += "|------------------------------|-------|\n"
		expected += "| \x1b[91mdotnetcore1.0（Deprecated）\x1b[0m  |     1 |\n"
		expected += "| \x1b[91mdotnetcore2.0（Deprecated）\x1b[0m  |     1 |\n"
		expected += "| dotnetcore2.1                |     1 |\n"
		expected += "| dotnetcore3.1                |     1 |\n"
		expected += "| go1.x                        |     4 |\n"
		expected += "| java11                       |     1 |\n"
		expected += "| java8                        |     1 |\n"
		expected += "| \x1b[91mnodejs（Deprecated）\x1b[0m         |     2 |\n"
		expected += "| nodejs10.x                   |     1 |\n"
		expected += "| nodejs12.x                   |     1 |\n"
		expected += "| \x1b[91mnodejs4.3（Deprecated）\x1b[0m      |     1 |\n"
		expected += "| \x1b[91mnodejs4.3-edge（Deprecated）\x1b[0m |     1 |\n"
		expected += "| \x1b[91mnodejs6.10（Deprecated）\x1b[0m     |     1 |\n"
		expected += "| \x1b[91mnodejs8.10（Deprecated）\x1b[0m     |     1 |\n"
		expected += "| provided                     |     6 |\n"
		expected += "| python3.6                    |     1 |\n"
		expected += "| python3.7                    |     1 |\n"
		expected += "| python3.8                    |     3 |\n"
		expected += "| ruby2.5                      |     1 |\n"
		expected += "| ruby2.7                      |     1 |\n"

		cmd := stats.NewCmd()
		var args []string
		data, err := stats.FetchData(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		actual := stats.Output(data)
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
		lm.AssertNumberOfCalls(t, "ListFunctions", 2)
	})
}
