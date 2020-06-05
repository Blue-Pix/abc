package ami_test

import (
	"fmt"
	"testing"

	"github.com/Blue-Pix/abc/lib/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
)

func TestToAMI(t *testing.T) {
	cases := []struct {
		in  *ssm.Parameter
		out ami.AMI
	}{
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs"),
				Value: aws.String("ami-0ff5dca93155f5191"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            false,
				Id:                 "ami-0ff5dca93155f5191",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2"),
				Value: aws.String("ami-0c3ae97724b825432"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "gp2",
				Minimal:            false,
				Id:                 "ami-0c3ae97724b825432",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-gp2",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3"),
				Value: aws.String("ami-03dd85055c8eb0ac9"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "s3",
				Minimal:            false,
				Id:                 "ami-03dd85055c8eb0ac9",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-hvm-x86_64-s3",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3"),
				Value: aws.String("ami-0e319478322617ce4"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "s3",
				Minimal:            true,
				Id:                 "ami-0e319478322617ce4",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-s3",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3"),
				Value: aws.String("ami-042998a62d60bf1ca"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "pv",
				Arch:               "x86_64",
				Storage:            "s3",
				Minimal:            true,
				Id:                 "ami-042998a62d60bf1ca",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-s3",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3"),
				Value: aws.String("ami-0a3e892fee3e3a614"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "pv",
				Arch:               "x86_64",
				Storage:            "s3",
				Minimal:            false,
				Id:                 "ami-0a3e892fee3e3a614",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-s3",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2"),
				Value: aws.String("ami-08360a37d07f61f88"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "2",
				VirtualizationType: "hvm",
				Arch:               "arm64",
				Storage:            "gp2",
				Minimal:            false,
				Id:                 "ami-08360a37d07f61f88",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-arm64-gp2",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs"),
				Value: aws.String("ami-06aa6ba9dc39dc071"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "2",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            false,
				Id:                 "ami-06aa6ba9dc39dc071",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"),
				Value: aws.String("ami-0f310fced6141e627"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "2",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "gp2",
				Minimal:            false,
				Id:                 "ami-0f310fced6141e627",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs"),
				Value: aws.String("ami-07fe0b0aed2e82d18"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "2",
				VirtualizationType: "hvm",
				Arch:               "arm64",
				Storage:            "ebs",
				Minimal:            true,
				Id:                 "ami-07fe0b0aed2e82d18",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-arm64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs"),
				Value: aws.String("ami-01bb806b33a3b98a3"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            true,
				Id:                 "ami-01bb806b33a3b98a3",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-hvm-x86_64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs"),
				Value: aws.String("ami-0690517f017a301c8"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "pv",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            true,
				Id:                 "ami-0690517f017a301c8",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-minimal-pv-x86_64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs"),
				Value: aws.String("ami-0c920068a5c30b361"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "1",
				VirtualizationType: "pv",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            false,
				Id:                 "ami-0c920068a5c30b361",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn-ami-pv-x86_64-ebs",
			},
		},
		{
			in: &ssm.Parameter{
				Name:  aws.String("/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs"),
				Value: aws.String("ami-03494c35f936e7fd7"),
				ARN:   aws.String("arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs"),
			},
			out: ami.AMI{
				Os:                 "amzn",
				Version:            "2",
				VirtualizationType: "hvm",
				Arch:               "x86_64",
				Storage:            "ebs",
				Minimal:            true,
				Id:                 "ami-03494c35f936e7fd7",
				Arn:                "arn:aws:ssm:ap-northeast-1::parameter/aws/service/ami-amazon-linux-latest/amzn2-ami-minimal-hvm-x86_64-ebs",
			},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			actual := ami.ToAMI(tt.in)
			assert.Equal(t, tt.out, actual)
		})
	}
}

func TestFetchData(t *testing.T) {
	type inputFlag struct {
		key, val string
	}

	cases := []struct {
		desc          string
		flags         []inputFlag
		expectedCount int
	}{
		{
			desc:          "without option",
			flags:         []inputFlag{},
			expectedCount: 14,
		},
		{
			desc: "with --version option",
			flags: []inputFlag{
				{key: "version", val: "2"},
			},
			expectedCount: 5,
		},
		{
			desc: "with --virtualization-type option",
			flags: []inputFlag{
				{key: "virtualization-type", val: "hvm"},
			},
			expectedCount: 10,
		},
		{
			desc: "with --arch option",
			flags: []inputFlag{
				{key: "arch", val: "x86_64"},
			},
			expectedCount: 12,
		},
		{
			desc: "with --storage option",
			flags: []inputFlag{
				{key: "storage", val: "gp2"},
			},
			expectedCount: 3,
		},
		{
			desc: "with --minimal option",
			flags: []inputFlag{
				{key: "minimal", val: "true"},
			},
			expectedCount: 6,
		},
		{
			desc: "with all options",
			flags: []inputFlag{
				{key: "version", val: "2"},
				{key: "virtualization-type", val: "hvm"},
				{key: "arch", val: "x86_64"},
				{key: "storage", val: "gp2"},
				{key: "minimal", val: "false"},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			cmd := ami.NewCmd()
			for _, flag := range tt.flags {
				cmd.Flags().Set(flag.key, flag.val)
			}
			sm := &ami.MockSsmClient{}
			ami.SetMockDefaultBehaviour(sm)
			ami.SsmClient = sm
			actual, err := ami.FetchData(cmd, []string{})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expectedCount, len(actual))
			assert.Nil(t, err)
			sm.AssertNumberOfCalls(t, "GetParametersByPath", 2)
		})
	}
}
