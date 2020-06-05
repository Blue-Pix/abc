package root

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Blue-Pix/abc/lib/ami"
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/purge_stack"
	"github.com/Blue-Pix/abc/lib/cfn/unused_exports"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/stretchr/testify/assert"
)

type mockCloudformationClient struct {
	cloudformationiface.CloudFormationAPI
}

func (client *mockCloudformationClient) ListStacks(params *cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error) {
	return &cloudformation.ListStacksOutput{
		NextToken: nil,
		StackSummaries: []*cloudformation.StackSummary{
			{StackId: aws.String("aaa"), StackName: aws.String("foo")},
			{StackId: aws.String("bbb"), StackName: aws.String("bar")},
			{StackId: aws.String("ccc"), StackName: aws.String("foobar")},
		},
	}, nil
}

func (client *mockCloudformationClient) ListExports(params *cloudformation.ListExportsInput) (*cloudformation.ListExportsOutput, error) {
	return &cloudformation.ListExportsOutput{
		NextToken: nil,
		Exports: []*cloudformation.Export{
			{Name: aws.String("foo_key1"), ExportingStackId: aws.String("aaa")},
			{Name: aws.String("foo_key2"), ExportingStackId: aws.String("aaa")},
			{Name: aws.String("bar_key1"), ExportingStackId: aws.String("bbb")},
			{Name: aws.String("bar_key2"), ExportingStackId: aws.String("bbb")},
		},
	}, nil
}

func (client *mockCloudformationClient) ListImports(params *cloudformation.ListImportsInput) (*cloudformation.ListImportsOutput, error) {
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

func TestExecute(t *testing.T) {
	t.Run("ami", func(t *testing.T) {
		t.Run("with shorthand options", func(t *testing.T) {
			args := []string{"ami", "-v", "2", "-V", "hvm", "-a", "x86_64", "-s", "gp2", "-m", "false"}
			cmd := NewCmd()
			cmd.SetArgs(args)
			amiCmd := ami.NewCmd()
			cmd.AddCommand(amiCmd)

			sm := &ami.MockSsmClient{}
			ami.SetMockDefaultBehaviour(sm)
			ami.SsmClient = sm

			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			expected := "[{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"}]\n"
			assert.Equal(t, expected, string(out))
			assert.Nil(t, err)
		})
	})

	t.Run("cfn", func(t *testing.T) {
		t.Run("unused-exports", func(t *testing.T) {
			t.Run("default", func(t *testing.T) {
				args := []string{"cfn", "unused-exports"}
				cmd := NewCmd()
				cmd.SetArgs(args)
				cfnCmd := cfn.NewCmd()
				unusedExportsCmd := unused_exports.NewCmd()
				cfnCmd.AddCommand(unusedExportsCmd)
				cmd.AddCommand(cfnCmd)
				unused_exports.Client = &mockCloudformationClient{}
				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				cmd.Execute()
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				expected := "[{\"name\":\"bar_key1\",\"exporting_stack\":\"bar\"},{\"name\":\"foo_key2\",\"exporting_stack\":\"foo\"}]\n"
				assert.Equal(t, expected, string(out))
				assert.Nil(t, err)
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
				assert.Nil(t, err)
			})
		})
	})
}
