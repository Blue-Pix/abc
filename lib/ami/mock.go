package ami

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/stretchr/testify/mock"
)

type MockSsmClient struct {
	mock.Mock
	ssmiface.SSMAPI
}

func (client *MockSsmClient) GetParametersByPath(params *ssm.GetParametersByPathInput) (*ssm.GetParametersByPathOutput, error) {
	args := client.Called(params)
	if args.Get(0) != nil {
		return args.Get(0).(*ssm.GetParametersByPathOutput), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

var MockData = [14]*ssm.Parameter{
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

func SetMockDefaultBehaviour(m *MockSsmClient) {
	m.On("GetParametersByPath", &ssm.GetParametersByPathInput{
		NextToken: nil,
		Path:      aws.String(PATH),
		Recursive: aws.Bool(true),
	}).Return(
		&ssm.GetParametersByPathOutput{
			NextToken:  aws.String("next_token"),
			Parameters: MockData[:7],
		},
		nil,
	)
	m.On("GetParametersByPath", &ssm.GetParametersByPathInput{
		NextToken: aws.String("next_token"),
		Path:      aws.String(PATH),
		Recursive: aws.Bool(true),
	}).Return(
		&ssm.GetParametersByPathOutput{
			NextToken:  nil,
			Parameters: MockData[7:],
		},
		nil,
	)
}
