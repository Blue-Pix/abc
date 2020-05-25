package ami

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	ssmiface.SSMAPI
}

func (client *mockClient) GetParametersByPath(params *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
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

func prepareCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := NewCmd()
	Client = &mockClient{}
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	return cmd, b
}

func TestMain(m *testing.M) {
	Client = &mockClient{}
	code := m.Run()
	os.Exit(code)
}

func TestRun(t *testing.T) {
	t.Run("without option", func(t *testing.T) {
		cmd, b := prepareCmd()
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-0ff5dca93155f5191\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0c3ae97724b825432\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":false,\"id\":\"ami-03dd85055c8eb0ac9\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-0e319478322617ce4\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-042998a62d60bf1ca\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":false,\"id\":\"ami-0a3e892fee3e3a614\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-08360a37d07f61f88\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-06aa6ba9dc39dc071\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-07fe0b0aed2e82d18\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-01bb806b33a3b98a3\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-0690517f017a301c8\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-0c920068a5c30b361\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-03494c35f936e7fd7\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("with --version option", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("version", "2")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-08360a37d07f61f88\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-06aa6ba9dc39dc071\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-07fe0b0aed2e82d18\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-03494c35f936e7fd7\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("with --virtualization-type option", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("virtualization-type", "hvm")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-0ff5dca93155f5191\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0c3ae97724b825432\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":false,\"id\":\"ami-03dd85055c8eb0ac9\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-0e319478322617ce4\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-08360a37d07f61f88\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-06aa6ba9dc39dc071\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-07fe0b0aed2e82d18\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-01bb806b33a3b98a3\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-03494c35f936e7fd7\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("with --arch option", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("arch", "x86_64")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-0ff5dca93155f5191\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0c3ae97724b825432\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":false,\"id\":\"ami-03dd85055c8eb0ac9\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-0e319478322617ce4\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-042998a62d60bf1ca\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":false,\"id\":\"ami-0a3e892fee3e3a614\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-06aa6ba9dc39dc071\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-01bb806b33a3b98a3\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-0690517f017a301c8\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":false,\"id\":\"ami-0c920068a5c30b361\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-03494c35f936e7fd7\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("with --storage option", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("storage", "gp2")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0c3ae97724b825432\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-08360a37d07f61f88\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("with --minimal option", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("minimal", "true")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-0e319478322617ce4\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"s3\",\"minimal\":true,\"id\":\"ami-042998a62d60bf1ca\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"arm64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-07fe0b0aed2e82d18\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-01bb806b33a3b98a3\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"1\",\"virtualization_type\":\"pv\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-0690517f017a301c8\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs\"},{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"ebs\",\"minimal\":true,\"id\":\"ami-03494c35f936e7fd7\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
	t.Run("withh all options", func(t *testing.T) {
		cmd, b := prepareCmd()
		cmd.Flags().Set("version", "2")
		cmd.Flags().Set("virtualization-type", "hvm")
		cmd.Flags().Set("arch", "x86_64")
		cmd.Flags().Set("storage", "gp2")
		cmd.Flags().Set("minimal", "false")
		err := cmd.Execute()
		out, err := ioutil.ReadAll(b)
		if err != nil {
			t.Fatal(err)
		}
		expected := "[{\"os\":\"amzn\",\"version\":\"2\",\"virtualization_type\":\"hvm\",\"arch\":\"x86_64\",\"storage\":\"gp2\",\"minimal\":false,\"id\":\"ami-0f310fced6141e627\",\"arn\":\"arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2\"}]\n"
		assert.Equal(t, expected, string(out))
		assert.Nil(t, err)
	})
}
