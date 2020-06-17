package create_stack_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/cfn/create_stack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
)

func initMockClient(cm *create_stack.MockCfnClient, sm *create_stack.MockS3Client) {
	create_stack.SetMockDefaultBehaviour(cm, sm)
	create_stack.CfnClient = cm
	create_stack.S3Client = sm
}

func TestExecCreateStack(t *testing.T) {
	cases := []struct {
		desc                        string
		stackName                   string
		templateInS3                string
		filePath                    string
		parameter1                  string
		parameter2                  string
		timeoutInMinutes            string
		notificationARNs            string
		capabilities                string
		roleArn                     string
		onFailure                   string
		tags                        string
		clientRequestToken          string
		enableTerminationProtection string
		createStackInput            *cloudformation.CreateStackInput
	}{
		{
			desc:                        "default",
			stackName:                   "test-stack",
			templateInS3:                "n",
			filePath:                    "../../../testdata/create-stack-sample.cf.yml",
			parameter1:                  "zzzzzzzz",
			parameter2:                  "yyyyyyyy",
			timeoutInMinutes:            "30",
			notificationARNs:            "a,b,c,d",
			capabilities:                "y",
			roleArn:                     "test-role",
			onFailure:                   "2",
			tags:                        "app=abc,env=test",
			clientRequestToken:          "test-token",
			enableTerminationProtection: "n",
			createStackInput: &cloudformation.CreateStackInput{
				StackName: aws.String("test-stack"),
				Parameters: []*cloudformation.Parameter{
					{
						ParameterKey:   aws.String("Param1"),
						ParameterValue: aws.String("zzzzzzzz"),
					},
					{
						ParameterKey:   aws.String("Param2"),
						ParameterValue: aws.String("yyyyyyyy"),
					},
				},
				Capabilities: []*string{
					aws.String("CAPABILITY_IAM"),
					aws.String("CAPABILITY_NAMED_IAM"),
					aws.String("CAPABILITY_AUTO_EXPAND"),
				},
				RoleARN:   aws.String("test-role"),
				OnFailure: aws.String("ROLLBACK"),
				NotificationARNs: []*string{
					aws.String("a"),
					aws.String("b"),
					aws.String("c"),
					aws.String("d"),
				},
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("app"),
						Value: aws.String("abc"),
					},
					{
						Key:   aws.String("env"),
						Value: aws.String("test"),
					},
				},
				ClientRequestToken:          aws.String("test-token"),
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(30)),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			f, _ := os.Open(tt.filePath)
			defer f.Close()
			b, _ := ioutil.ReadAll(f)
			tt.createStackInput.SetTemplateBody(string(b))

			cm := &create_stack.MockCfnClient{}
			cm.On("CreateStack", tt.createStackInput).Return(
				&cloudformation.CreateStackOutput{
					StackId: aws.String("1234567"),
				},
				nil,
			)
			sm := &create_stack.MockS3Client{}
			initMockClient(cm, sm)

			cmd := create_stack.NewCmd()
			input := []string{
				tt.stackName,
				tt.templateInS3,
				tt.filePath,
				tt.parameter1,
				tt.parameter2,
				tt.timeoutInMinutes,
				tt.notificationARNs,
				tt.capabilities,
				tt.roleArn,
				tt.onFailure,
				tt.tags,
				tt.clientRequestToken,
				tt.enableTerminationProtection,
			}
			cmd.SetIn(bytes.NewBufferString(strings.Join(input, "\n")))
			stackId, err := create_stack.ExecCreateStack(cmd, []string{})

			assert.Equal(t, "1234567", stackId)
			assert.Nil(t, err)
			cm.AssertNumberOfCalls(t, "CreateStack", 1)
			sm.AssertNumberOfCalls(t, "GetObject", 0)
		})
	}
}
