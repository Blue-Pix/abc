package root

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/ami"
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/unused_exports"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockSSMClient struct {
	ssmiface.SSMAPI
}

func (client *mockSSMClient) GetParametersByPath(params *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return &ssm.GetParametersByPathOutput{
		NextToken: nil,
		Parameters: []*ssm.Parameter{
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs"), Value: aws.String("ami-0ff5dca93155f5191"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2"), Value: aws.String("ami-0c3ae97724b825432"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3"), Value: aws.String("ami-03dd85055c8eb0ac9"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3"), Value: aws.String("ami-0e319478322617ce4"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3"), Value: aws.String("ami-042998a62d60bf1ca"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3"), Value: aws.String("ami-0a3e892fee3e3a614"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2"), Value: aws.String("ami-08360a37d07f61f88"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs"), Value: aws.String("ami-06aa6ba9dc39dc071"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"), Value: aws.String("ami-0f310fced6141e627"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs"), Value: aws.String("ami-07fe0b0aed2e82d18"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs"), Value: aws.String("ami-01bb806b33a3b98a3"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs"), Value: aws.String("ami-0690517f017a301c8"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs"), Value: aws.String("ami-0c920068a5c30b361"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs")},
			{Name: aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs"), Value: aws.String("ami-03494c35f936e7fd7"), ARN: aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs")},
		},
	}, nil
}

func prepareCfnCmd(args []string) *cobra.Command {
	cmd := NewCmd()
	cmd.SetArgs(args)
	cfnCmd := cfn.NewCmd()
	unusedExportsCmd := unused_exports.NewCmd()
	cfnCmd.AddCommand(unusedExportsCmd)
	cmd.AddCommand(cfnCmd)
	return cmd
}

func TestExecute(t *testing.T) {
	t.Run("ami", func(t *testing.T) {
		t.Run("with shorthand options", func(t *testing.T) {
			args := []string{"ami", "-v", "2", "-V", "hvm", "-a", "x86_64", "-s", "gp2", "-m", "false"}
			cmd := NewCmd()
			cmd.SetArgs(args)
			amiCmd := ami.NewCmd()
			cmd.AddCommand(amiCmd)
			ami.Client = &mockSSMClient{}
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
				cmd := prepareCfnCmd(args)
				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				cmd.Execute()
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				list := strings.Split(string(out), "\n")
				assert.Regexp(t, regexp.MustCompile(`^\.+$`), list[0])
				assert.Equal(t, "name,exporting_stack", list[1])
				assert.Regexp(t, regexp.MustCompile(`^.+,.+$`), list[2])
			})
		})
	})
}
