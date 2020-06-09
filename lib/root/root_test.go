package root

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/ami"
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/purge_stack"
	"github.com/Blue-Pix/abc/lib/cfn/unused_exports"
	"github.com/Blue-Pix/abc/lib/lambda"
	"github.com/Blue-Pix/abc/lib/lambda/stats"
	awsLambda "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecute(t *testing.T) {
	t.Run("ami", func(t *testing.T) {
		sm := &ami.MockSsmClient{}
		ami.SetMockDefaultBehaviour(sm)
		ami.SsmClient = sm
		expected := "[{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"}]\n"

		t.Run("success with shorthand options", func(t *testing.T) {
			args := []string{"ami", "-v", "2", "-V", "hvm", "-a", "x86_64", "-s", "gp2", "-m", "false"}
			cmd := NewCmd()
			cmd.SetArgs(args)
			amiCmd := ami.NewCmd()
			cmd.AddCommand(amiCmd)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, expected, string(out))
		})

		t.Run("success with long options", func(t *testing.T) {
			args := []string{"ami", "--version", "2", "--virtualization-type", "hvm", "--arch", "x86_64", "--storage", "gp2", "--minimal", "false"}
			cmd := NewCmd()
			cmd.SetArgs(args)
			amiCmd := ami.NewCmd()
			cmd.AddCommand(amiCmd)

			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, expected, string(out))
		})
	})

	t.Run("cfn", func(t *testing.T) {
		t.Run("unused-exports", func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				args := []string{"cfn", "unused-exports"}
				cmd := NewCmd()
				cmd.SetArgs(args)
				cfnCmd := cfn.NewCmd()
				unusedExportsCmd := unused_exports.NewCmd()
				cfnCmd.AddCommand(unusedExportsCmd)
				cmd.AddCommand(cfnCmd)

				cm := &unused_exports.MockCfnClient{}
				unused_exports.SetMockDefaultBehaviour(cm)
				unused_exports.CfnClient = cm

				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				cmd.Execute()
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				expected := "[{\"name\":\"bar_key1\",\"exporting_stack\":\"bar\"},{\"name\":\"foo_key2\",\"exporting_stack\":\"foo\"}]\n"
				assert.Equal(t, expected, string(out))
			})
		})

		t.Run("purge-stack", func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				stackName := "foo"
				args := []string{"cfn", "purge-stack", "--stack-name", stackName}
				cmd := NewCmd()
				cmd.SetArgs(args)
				cfnCmd := cfn.NewCmd()
				purgeStackCmd := purge_stack.NewCmd()
				cfnCmd.AddCommand(purgeStackCmd)
				cmd.AddCommand(cfnCmd)

				cm := &purge_stack.MockCfnClient{}
				em := &purge_stack.MockEcrClient{}
				purge_stack.SetMockDefaultBehaviour(cm, em)
				purge_stack.CfnClient = cm
				purge_stack.EcrClient = em

				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				cmd.Execute()
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				expected := "All images in ecr1 successfully deleted.\nPerform delete-stack is in progress asynchronously.\nPlease check deletion status by yourself.\n"
				assert.Equal(t, expected, string(out))
			})
		})
	})

	t.Run("lambda", func(t *testing.T) {
		t.Run("stats", func(t *testing.T) {
			t.Run("table format", func(t *testing.T) {
				t.Run("with no verbose option", func(t *testing.T) {
					args := []string{"lambda", "stats"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "|           RUNTIME            | COUNT |\n"
					assert.True(t, strings.HasPrefix(string(out), expected))
				})

				t.Run("with verbose option", func(t *testing.T) {
					args := []string{"lambda", "stats", "--verbose"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "|           RUNTIME            | COUNT |           FUNCTIONS            |\n"
					assert.True(t, strings.HasPrefix(string(out), expected))
				})

				t.Run("no function exists", func(t *testing.T) {
					args := []string{"lambda", "stats"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					lm.On("ListFunctions", mock.AnythingOfType("*lambda.ListFunctionsInput")).Return(&awsLambda.ListFunctionsOutput{
						NextMarker: nil,
						Functions:  []*awsLambda.FunctionConfiguration{},
					}, nil)
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "no function found.\n"
					assert.Equal(t, expected, string(out))
				})
			})

			t.Run("json format", func(t *testing.T) {
				t.Run("with no verbose option", func(t *testing.T) {
					args := []string{"lambda", "stats", "--format", "json"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "[{\"runtime\":\"dotnetcore1.0\",\"count\":1,\"deprecated\":true},"
					assert.True(t, strings.HasPrefix(string(out), expected))
				})

				t.Run("with verbose option", func(t *testing.T) {
					args := []string{"lambda", "stats", "--format", "json", "--verbose"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "[{\"runtime\":\"dotnetcore1.0\",\"count\":1,\"functions\":[\"dotnet1.0-func-1\"],\"deprecated\":true},"
					assert.True(t, strings.HasPrefix(string(out), expected))
				})

				t.Run("no function exists", func(t *testing.T) {
					args := []string{"lambda", "stats", "--format", "json"}
					cmd := NewCmd()
					cmd.SetArgs(args)
					lambdaCmd := lambda.NewCmd()
					statsCmd := stats.NewCmd()
					lambdaCmd.AddCommand(statsCmd)
					cmd.AddCommand(lambdaCmd)

					lm := &stats.MockLambdaClient{}
					lm.On("ListFunctions", mock.AnythingOfType("*lambda.ListFunctionsInput")).Return(&awsLambda.ListFunctionsOutput{
						NextMarker: nil,
						Functions:  []*awsLambda.FunctionConfiguration{},
					}, nil)
					stats.SetMockDefaultBehaviour(lm)
					stats.LambdaClient = lm

					b := bytes.NewBufferString("")
					cmd.SetOut(b)
					cmd.Execute()
					out, err := ioutil.ReadAll(b)
					if err != nil {
						t.Fatal(err)
					}
					expected := "[]\n"
					assert.Equal(t, expected, string(out))
				})
			})

			t.Run("invalid format", func(t *testing.T) {
				args := []string{"lambda", "stats", "--format", "hoge"}
				cmd := NewCmd()
				cmd.SetArgs(args)
				lambdaCmd := lambda.NewCmd()
				statsCmd := stats.NewCmd()
				lambdaCmd.AddCommand(statsCmd)
				cmd.AddCommand(lambdaCmd)

				lm := &stats.MockLambdaClient{}
				stats.SetMockDefaultBehaviour(lm)
				stats.LambdaClient = lm

				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				err := cmd.Execute()
				assert.Error(t, err, "invalid format.")
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				expected := ""
				assert.Equal(t, expected, string(out))
			})
		})
	})
}
