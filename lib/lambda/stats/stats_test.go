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
	t.Run("table format", func(t *testing.T) {
		t.Run("with no verbose option", func(t *testing.T) {
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
			cmd.Flags().Set("format", "table")
			var args []string
			data, err := stats.FetchData(cmd, args)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := stats.Output(data)
			assert.Equal(t, expected, actual)
			assert.Nil(t, err)
			lm.AssertNumberOfCalls(t, "ListFunctions", 2)
		})

		t.Run("with verbose option", func(t *testing.T) {
			lm := &stats.MockLambdaClient{}
			initMockClient(lm)

			expected := "|           RUNTIME            | COUNT |           FUNCTIONS            |\n"
			expected += "|------------------------------|-------|--------------------------------|\n"
			expected += "| \x1b[91mdotnetcore1.0（Deprecated）\x1b[0m  |     1 | dotnet1.0-func-1               |\n"
			expected += "| \x1b[91mdotnetcore2.0（Deprecated）\x1b[0m  |     1 | dotnet2.0-func-1               |\n"
			expected += "| dotnetcore2.1                |     1 | dotnet2.1-func-1               |\n"
			expected += "| dotnetcore3.1                |     1 | dotnet3.1-func-1               |\n"
			expected += "| go1.x                        |     4 | go1-func-1, go1-func-2,        |\n"
			expected += "|                              |       | go1-func-3, go1-func-4         |\n"
			expected += "| java11                       |     1 | java11-func-1                  |\n"
			expected += "| java8                        |     1 | java8-func-1                   |\n"
			expected += "| \x1b[91mnodejs（Deprecated）\x1b[0m         |     2 | nodejs0.10-func-1,             |\n"
			expected += "|                              |       | nodejs0.10-func-2              |\n"
			expected += "| nodejs10.x                   |     1 | node10-func-1                  |\n"
			expected += "| nodejs12.x                   |     1 | node12-func-1                  |\n"
			expected += "| \x1b[91mnodejs4.3（Deprecated）\x1b[0m      |     1 | nodejs4.3-func-1               |\n"
			expected += "| \x1b[91mnodejs4.3-edge（Deprecated）\x1b[0m |     1 | nodejs4.3edge-func-1           |\n"
			expected += "| \x1b[91mnodejs6.10（Deprecated）\x1b[0m     |     1 | nodejs6.10-func-1              |\n"
			expected += "| \x1b[91mnodejs8.10（Deprecated）\x1b[0m     |     1 | nodejs8.10-func-1              |\n"
			expected += "| provided                     |     6 | provided-func-1,               |\n"
			expected += "|                              |       | provided-func-2,               |\n"
			expected += "|                              |       | provided-func-3,               |\n"
			expected += "|                              |       | provided-func-4,               |\n"
			expected += "|                              |       | provided-func-5,               |\n"
			expected += "|                              |       | provided-func-6                |\n"
			expected += "| python3.6                    |     1 | python3.6-func-1               |\n"
			expected += "| python3.7                    |     1 | python3.7-func-1               |\n"
			expected += "| python3.8                    |     3 | python3.8-func-1,              |\n"
			expected += "|                              |       | python3.8-func-2,              |\n"
			expected += "|                              |       | python3.8-func-3               |\n"
			expected += "| ruby2.5                      |     1 | ruby2.5-func-1                 |\n"
			expected += "| ruby2.7                      |     1 | ruby2.7-func-1                 |\n"

			cmd := stats.NewCmd()
			cmd.Flags().Set("format", "table")
			cmd.Flags().Set("verbose", "true")
			var args []string
			data, err := stats.FetchData(cmd, args)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := stats.Output(data)
			assert.Equal(t, expected, actual)
			assert.Nil(t, err)
			lm.AssertNumberOfCalls(t, "ListFunctions", 2)
		})
	})

	t.Run("json format", func(t *testing.T) {
		t.Run("with no verbose option", func(t *testing.T) {
			lm := &stats.MockLambdaClient{}
			initMockClient(lm)

			expected := "["
			expected += "{\"runtime\":\"dotnetcore1.0\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"dotnetcore2.0\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"dotnetcore2.1\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"dotnetcore3.1\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"go1.x\",\"count\":4,\"deprecated\":false},"
			expected += "{\"runtime\":\"java11\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"java8\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs\",\"count\":2,\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs10.x\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs12.x\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs4.3\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs4.3-edge\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs6.10\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs8.10\",\"count\":1,\"deprecated\":true},"
			expected += "{\"runtime\":\"provided\",\"count\":6,\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.6\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.7\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.8\",\"count\":3,\"deprecated\":false},"
			expected += "{\"runtime\":\"ruby2.5\",\"count\":1,\"deprecated\":false},"
			expected += "{\"runtime\":\"ruby2.7\",\"count\":1,\"deprecated\":false}"
			expected += "]"

			cmd := stats.NewCmd()
			cmd.Flags().Set("format", "json")
			var args []string
			data, err := stats.FetchData(cmd, args)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := stats.Output(data)
			assert.Equal(t, expected, actual)
			assert.Nil(t, err)
			lm.AssertNumberOfCalls(t, "ListFunctions", 2)
		})

		t.Run("with verbose option", func(t *testing.T) {
			lm := &stats.MockLambdaClient{}
			initMockClient(lm)

			expected := "["
			expected += "{\"runtime\":\"dotnetcore1.0\",\"count\":1,\"functions\":[\"dotnet1.0-func-1\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"dotnetcore2.0\",\"count\":1,\"functions\":[\"dotnet2.0-func-1\"],"
			expected += "\"deprecated\":true},{\"runtime\":\"dotnetcore2.1\",\"count\":1,\"functions\":[\"dotnet2.1-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"dotnetcore3.1\",\"count\":1,\"functions\":[\"dotnet3.1-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"go1.x\",\"count\":4,\"functions\":[\"go1-func-1\",\"go1-func-2\",\"go1-func-3\",\"go1-func-4\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"java11\",\"count\":1,\"functions\":[\"java11-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"java8\",\"count\":1,\"functions\":[\"java8-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs\",\"count\":2,\"functions\":[\"nodejs0.10-func-1\",\"nodejs0.10-func-2\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs10.x\",\"count\":1,\"functions\":[\"node10-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs12.x\",\"count\":1,\"functions\":[\"node12-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"nodejs4.3\",\"count\":1,\"functions\":[\"nodejs4.3-func-1\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs4.3-edge\",\"count\":1,\"functions\":[\"nodejs4.3edge-func-1\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs6.10\",\"count\":1,\"functions\":[\"nodejs6.10-func-1\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"nodejs8.10\",\"count\":1,\"functions\":[\"nodejs8.10-func-1\"],\"deprecated\":true},"
			expected += "{\"runtime\":\"provided\",\"count\":6,\"functions\":[\"provided-func-1\",\"provided-func-2\",\"provided-func-3\",\"provided-func-4\",\"provided-func-5\",\"provided-func-6\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.6\",\"count\":1,\"functions\":[\"python3.6-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.7\",\"count\":1,\"functions\":[\"python3.7-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"python3.8\",\"count\":3,\"functions\":[\"python3.8-func-1\",\"python3.8-func-2\",\"python3.8-func-3\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"ruby2.5\",\"count\":1,\"functions\":[\"ruby2.5-func-1\"],\"deprecated\":false},"
			expected += "{\"runtime\":\"ruby2.7\",\"count\":1,\"functions\":[\"ruby2.7-func-1\"],\"deprecated\":false}"
			expected += "]"

			cmd := stats.NewCmd()
			cmd.Flags().Set("format", "json")
			cmd.Flags().Set("verbose", "true")
			var args []string
			data, err := stats.FetchData(cmd, args)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := stats.Output(data)
			assert.Equal(t, expected, actual)
			assert.Nil(t, err)
			lm.AssertNumberOfCalls(t, "ListFunctions", 2)
		})
	})

	t.Run("invalid format", func(t *testing.T) {
		lm := &stats.MockLambdaClient{}
		initMockClient(lm)

		expected := ""
		cmd := stats.NewCmd()
		cmd.Flags().Set("format", "hoge")
		var args []string
		data, err := stats.FetchData(cmd, args)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := stats.Output(data)
		assert.Equal(t, expected, actual)
		assert.EqualError(t, err, "invalid format.")
		lm.AssertNumberOfCalls(t, "ListFunctions", 2)
	})
}
