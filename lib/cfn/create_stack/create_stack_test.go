package create_stack_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/cfn/create_stack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func initMockClient(cm *create_stack.MockCfnClient, sm *create_stack.MockS3Client) {
	create_stack.SetMockDefaultBehaviour(cm, sm)
	create_stack.CfnClient = cm
	create_stack.S3Client = sm
}

type testCase struct {
	desc                        string
	stackName                   string // user input
	templateInS3                string // user input
	filePath                    string // user input
	parameter1                  string // user input
	parameter2                  string // user input
	timeoutInMinutes            string // user input
	notificationARNs            string // user input
	capabilities                string // user input
	roleArn                     string // user input
	onFailure                   string // user input
	tags                        string // user input
	clientRequestToken          string // user input
	enableTerminationProtection string // user input
	createStackInput            *cloudformation.CreateStackInput
	expected_error              error
}

func makeInputBuffer(c testCase) *bytes.Buffer {
	input := []string{
		c.stackName,
		c.templateInS3,
		c.filePath,
		c.parameter1,
		c.parameter2,
		c.timeoutInMinutes,
		c.notificationARNs,
		c.capabilities,
		c.roleArn,
		c.onFailure,
		c.tags,
		c.clientRequestToken,
		c.enableTerminationProtection,
	}
	return bytes.NewBufferString(strings.Join(input, "\n"))
}

func readOutput(t *testing.T, b io.Reader) string {
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func compileRegex(t *testing.T, exp string) *regexp.Regexp {
	reg, err := regexp.Compile(exp)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestExecCreateStack(t *testing.T) {
	green_cases := []testCase{
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

	for _, tt := range green_cases {
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
			cmd.SetIn(makeInputBuffer(tt))
			stackId, err := create_stack.ExecCreateStack(cmd, []string{})

			assert.Equal(t, "1234567", stackId)
			assert.Nil(t, err)
			cm.AssertNumberOfCalls(t, "CreateStack", 1)
			sm.AssertNumberOfCalls(t, "GetObject", 0)
		})
	}

	red_cases := []testCase{
		{
			desc:           "stack name is empty",
			stackName:      "\nhoge",
			templateInS3:   "n",
			filePath:       "./invalid_path",
			expected_error: errors.New("There was an error processing the template: open ./invalid_path: no such file or directory"),
		},
	}

	for _, tt := range red_cases {
		t.Run(tt.desc, func(t *testing.T) {
			cm := &create_stack.MockCfnClient{}
			cm.On("CreateStack", mock.AnythingOfType("*cloudformation.CreateStackInput")).Return(
				&cloudformation.CreateStackOutput{
					StackId: aws.String("1234567"),
				},
				nil,
			)
			sm := &create_stack.MockS3Client{}
			initMockClient(cm, sm)

			cmd := create_stack.NewCmd()
			cmd.SetIn(makeInputBuffer(tt))
			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			stackId, err := create_stack.ExecCreateStack(cmd, []string{})

			expectedOutput := compileRegex(t, "Stack name:.+Stack name:.+")
			actualOutput := readOutput(t, b)
			assert.Regexp(t, expectedOutput, actualOutput)
			assert.Equal(t, "", stackId)
			assert.Equal(t, tt.expected_error, err)
			cm.AssertNumberOfCalls(t, "CreateStack", 0)
			sm.AssertNumberOfCalls(t, "GetObject", 0)
		})
	}
}
