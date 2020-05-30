package ami

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	ssmiface.SSMAPI
}

var test_parameters = []*ssm.Parameter{
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
}

func (client *mockClient) GetParametersByPath(params *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	return &ssm.GetParametersByPathOutput{
		NextToken:  nil,
		Parameters: test_parameters,
	}, nil
}

var test_amis []AMI

func TestMain(m *testing.M) {
	Client = &mockClient{}
	for _, p := range test_parameters {
		test_amis = append(test_amis, ToAMI(p))
	}
	code := m.Run()
	os.Exit(code)
}

func TestFetchData(t *testing.T) {
	t.Run("without option", func(t *testing.T) {
		cmd := NewCmd()
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := test_amis
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("with --version option", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("version", "2")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[6], test_amis[7], test_amis[8], test_amis[9], test_amis[13]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("with --virtualization-type option", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("virtualization-type", "hvm")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[0], test_amis[1], test_amis[2], test_amis[3], test_amis[6], test_amis[7], test_amis[8], test_amis[9], test_amis[10], test_amis[13]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("with --arch option", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("arch", "x86_64")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[0], test_amis[1], test_amis[2], test_amis[3], test_amis[4], test_amis[5], test_amis[7], test_amis[8], test_amis[10], test_amis[11], test_amis[12], test_amis[13]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("with --storage option", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("storage", "gp2")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[1], test_amis[6], test_amis[8]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("with --minimal option", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("minimal", "true")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[3], test_amis[4], test_amis[9], test_amis[10], test_amis[11], test_amis[13]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
	t.Run("withh all options", func(t *testing.T) {
		cmd := NewCmd()
		cmd.Flags().Set("version", "2")
		cmd.Flags().Set("virtualization-type", "hvm")
		cmd.Flags().Set("arch", "x86_64")
		cmd.Flags().Set("storage", "gp2")
		cmd.Flags().Set("minimal", "false")
		actual, err := FetchData(cmd, []string{})
		if err != nil {
			t.Fatal(err)
		}
		expected := []AMI{test_amis[8]}
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
}
