package create_stack_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/cfn/create_stack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const invalidSampleLocalTemplateFilePath = "../../../testdata/create_stack/invalid.cf.yml"
const noParamsSampleLocalTemplateFilePath = "../../../testdata/create_stack/no-params.cf.yml"
const withParamsTemplateFilePath = "../../../testdata/create_stack/with-params.cf.yml"

func initMockClient(cm *create_stack.MockCfnClient, sm *create_stack.MockS3Client) {
	create_stack.SetMockDefaultBehaviour(cm, sm)
	create_stack.CfnClient = cm
	create_stack.S3Client = sm
}

type userInput struct {
	StackName                   string
	TemplateInS3                string
	FilePath                    string
	BucketName                  string
	BucketKey                   string
	BucketRegion                string
	Parameters                  string
	TimeoutInMinutes            string
	NotificationARNs            string
	Capabilities                string
	RoleArn                     string
	OnFailure                   string
	Tags                        string
	ClientRequestToken          string
	EnableTerminationProtection string
}

func initializeUserInput() userInput {
	return userInput{
		StackName:                   "sample-stack-name",
		TemplateInS3:                create_stack.NO,
		FilePath:                    noParamsSampleLocalTemplateFilePath,
		BucketName:                  "sample-bucket",
		BucketKey:                   "sample-key",
		BucketRegion:                "us-west-2",
		TimeoutInMinutes:            strconv.Itoa(create_stack.DefaultTimeoutInMinutes),
		NotificationARNs:            "a,b,c,d",
		Capabilities:                create_stack.YES,
		RoleArn:                     "sample-role",
		OnFailure:                   create_stack.ROLLBACK,
		Tags:                        "sample=tag,foo=bar",
		ClientRequestToken:          "sample-request-token",
		EnableTerminationProtection: create_stack.NO,
	}
}

func (ui *userInput) overwriteUserInput(fieldName string, value string) {
	reflect.ValueOf(ui).Elem().FieldByName(fieldName).SetString(value)
}

func (ui userInput) buildTemplateUrl() string {
	return fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", ui.BucketName, ui.BucketRegion, ui.BucketKey)
}

func (ui userInput) buildTemplateBody() string {
	f, _ := os.Open(ui.FilePath)
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	return string(b)
}

func (ui userInput) makeUserInputBuffer() *bytes.Buffer {
	input := []string{
		ui.StackName,
		ui.TemplateInS3,
	}
	if ui.TemplateInS3 == create_stack.YES {
		input = append(input, []string{
			ui.BucketName,
			ui.BucketKey,
			ui.BucketRegion,
		}...)
	} else {
		input = append(input, ui.FilePath)
	}
	if len(ui.Parameters) > 0 {
		input = append(input, ui.Parameters)
	}
	input = append(input, []string{
		ui.TimeoutInMinutes,
		ui.NotificationARNs,
		ui.Capabilities,
		ui.RoleArn,
		ui.OnFailure,
		ui.Tags,
		ui.ClientRequestToken,
		ui.EnableTerminationProtection,
	}...)
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
	green_cases := []struct {
		desc                     string
		userInput                userInput
		expectedCreateStackInput *cloudformation.CreateStackInput
		getObjectCalledCount     int
	}{
		{
			desc: "template file located in local",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "template file located in S3",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.YES,
				BucketName:                  "sample-bucket",
				BucketKey:                   noParamsSampleLocalTemplateFilePath,
				BucketRegion:                "us-west2",
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
			getObjectCalledCount: 1,
		},
		{
			desc: "template contains no params",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "use default value for params",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    withParamsTemplateFilePath,
				Parameters:                  "\n\n\n",
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName: aws.String("test-stack"),
				Parameters: []*cloudformation.Parameter{
					{
						ParameterKey:   aws.String("ParamWithDefaultValue"),
						ParameterValue: aws.String("this is default value for ParamWithDefaultValue"),
					},
				},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "overwrite default value for params",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    withParamsTemplateFilePath,
				Parameters:                  "foo\nbar\nfizz\nbuzz",
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName: aws.String("test-stack"),
				Parameters: []*cloudformation.Parameter{
					{
						ParameterKey:   aws.String("ParamWithDefaultValue"),
						ParameterValue: aws.String("foo"),
					},
					{
						ParameterKey:   aws.String("ParamWithDescription"),
						ParameterValue: aws.String("bar"),
					},
					{
						ParameterKey:   aws.String("ParamWithoutDefaultValue"),
						ParameterValue: aws.String("fizz"),
					},
					{
						ParameterKey:   aws.String("ParamWithoutDescription"),
						ParameterValue: aws.String("buzz"),
					},
				},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "timeoutInMinutes is empty",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "timeoutInMinutes exists",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "30",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(30),
			},
		},
		{
			desc: "notificationARNs is empty",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "notificationARNs exist",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "aaa,bbb,ccc",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:    aws.String("test-stack"),
				Parameters:   []*cloudformation.Parameter{},
				Capabilities: create_stack.CAPABILITIES,
				OnFailure:    aws.String("ROLLBACK"),
				NotificationARNs: []*string{
					aws.String("aaa"),
					aws.String("bbb"),
					aws.String("ccc"),
				},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "acknowledge to pass capabilities",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "roleArn is empty",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "roleArn exists",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "sample-role-arn",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				RoleARN:                     aws.String("sample-role-arn"),
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "DO_NOTHING onFailure",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.DO_NOTHING,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("DO_NOTHING"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "ROLLBACK onFailure",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "DELETE onFailure",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.DELETE,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("DELETE"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "tags is empty",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "tags exists",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "app=abc,env=staging",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:        aws.String("test-stack"),
				Parameters:       []*cloudformation.Parameter{},
				Capabilities:     create_stack.CAPABILITIES,
				OnFailure:        aws.String("ROLLBACK"),
				NotificationARNs: []*string{},
				Tags: []*cloudformation.Tag{
					{
						Key:   aws.String("app"),
						Value: aws.String("abc"),
					},
					{
						Key:   aws.String("env"),
						Value: aws.String("staging"),
					},
				},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "clientRequestToken is empty",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "clientRequestToken exists",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "sample-token",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				ClientRequestToken:          aws.String("sample-token"),
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "enableTerminationProtection is true",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.YES,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(true),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
		{
			desc: "enableTerminationProtection is false",
			userInput: userInput{
				StackName:                   "test-stack",
				TemplateInS3:                create_stack.NO,
				FilePath:                    noParamsSampleLocalTemplateFilePath,
				TimeoutInMinutes:            "",
				NotificationARNs:            "",
				Capabilities:                create_stack.YES,
				RoleArn:                     "",
				OnFailure:                   create_stack.ROLLBACK,
				Tags:                        "",
				ClientRequestToken:          "",
				EnableTerminationProtection: create_stack.NO,
			},
			expectedCreateStackInput: &cloudformation.CreateStackInput{
				StackName:                   aws.String("test-stack"),
				Parameters:                  []*cloudformation.Parameter{},
				Capabilities:                create_stack.CAPABILITIES,
				OnFailure:                   aws.String("ROLLBACK"),
				NotificationARNs:            []*string{},
				Tags:                        []*cloudformation.Tag{},
				EnableTerminationProtection: aws.Bool(false),
				TimeoutInMinutes:            aws.Int64(int64(create_stack.DefaultTimeoutInMinutes)),
			},
		},
	}

	for _, tt := range green_cases {
		t.Run(fmt.Sprintf("success/%s", tt.desc), func(t *testing.T) {
			if tt.userInput.TemplateInS3 == create_stack.YES {
				tt.expectedCreateStackInput.SetTemplateURL(tt.userInput.buildTemplateUrl())
			} else {
				tt.expectedCreateStackInput.SetTemplateBody(tt.userInput.buildTemplateBody())
			}

			cm := &create_stack.MockCfnClient{}
			cm.On("CreateStack", tt.expectedCreateStackInput).Return(
				&cloudformation.CreateStackOutput{
					StackId: aws.String("1234567"),
				},
				nil,
			)
			sm := &create_stack.MockS3Client{}
			initMockClient(cm, sm)

			cmd := create_stack.NewCmd()
			cmd.SetIn(tt.userInput.makeUserInputBuffer())
			stackId, err := create_stack.ExecCreateStack(cmd, []string{})

			assert.Equal(t, "1234567", stackId)
			assert.Nil(t, err)
			cm.AssertNumberOfCalls(t, "CreateStack", 1)
			sm.AssertNumberOfCalls(t, "GetObject", tt.getObjectCalledCount)
		})
	}

	params_cases := []struct {
		desc                string
		inputDefinitions    map[string]string
		expectedOutputRegex string
	}{
		{
			desc: "check parameters question format",
			inputDefinitions: map[string]string{
				"FilePath":   withParamsTemplateFilePath,
				"Parameters": "foo\nbar\nfizz\nbuzz",
			},
			expectedOutputRegex: `(?s)ParamWithDefaultValue \[this is default value for ParamWithDefaultValue\]:  ParamWithDescription \(this is description for ParamWithDescription\):  ParamWithoutDefaultValue:  ParamWithoutDescription:`,
		},
	}

	for _, tt := range params_cases {
		t.Run(fmt.Sprintf("params_check/%s", tt.desc), func(t *testing.T) {
			cm := &create_stack.MockCfnClient{}
			sm := &create_stack.MockS3Client{}
			initMockClient(cm, sm)

			cmd := create_stack.NewCmd()

			ui := initializeUserInput()
			for k, v := range tt.inputDefinitions {
				ui.overwriteUserInput(k, v)
			}
			cmd.SetIn(ui.makeUserInputBuffer())
			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			_, _ = create_stack.ExecCreateStack(cmd, []string{})

			expectedOutput := compileRegex(t, tt.expectedOutputRegex)
			actualOutput := readOutput(t, b)
			assert.Regexp(t, expectedOutput, actualOutput)
		})
	}

	// check to asked repeatedly
	invalid_input_cases := []struct {
		desc                string
		inputDefinitions    map[string]string
		expectedOutputRegex string
	}{
		{
			desc: "stack name is empty",
			inputDefinitions: map[string]string{
				"StackName": "\nhoge",
			},
			expectedOutputRegex: "Stack name:.+Stack name:.+",
		},
		{
			desc: "invalid input for template location",
			inputDefinitions: map[string]string{
				"TemplateInS3": fmt.Sprintf("invalid_string\n%s", create_stack.NO),
			},
			expectedOutputRegex: "Template in S3?.+Template in S3?.+",
		},
		{
			desc: "bucket name is empty",
			inputDefinitions: map[string]string{
				"TemplateInS3": create_stack.YES,
				"BucketName":   "\nsample-bucket",
				"BucketKey":    "sample-bucket-key",
				"BucketRegion": "us-west2",
			},
		},
		{
			desc: "bucket key is empty",
			inputDefinitions: map[string]string{
				"TemplateInS3": create_stack.YES,
				"BucketName":   "sample-bucket",
				"BucketKey":    "\nsample-bucket-key",
				"BucketRegion": "us-west2",
			},
			expectedOutputRegex: "S3 Bucket key:.+S3 Bucket key:.+",
		},
		{
			desc: "bucket region is empty",
			inputDefinitions: map[string]string{
				"TemplateInS3": create_stack.YES,
				"BucketName":   "sample-bucket",
				"BucketKey":    "sample-bucket-key",
				"BucketRegion": "\nus-west2",
			},
			expectedOutputRegex: "S3 Bucket region:.+S3 Bucket region:.+",
		},
		{
			desc: "file path is empty",
			inputDefinitions: map[string]string{
				"FilePath": fmt.Sprintf("\n%s", noParamsSampleLocalTemplateFilePath),
			},
			expectedOutputRegex: "File path:.+File path:.+",
		},
		{
			desc: "timeoutInMinutes is not integer",
			inputDefinitions: map[string]string{
				"TimeoutInMinutes": "not_int_input\n30",
			},
			expectedOutputRegex: "Timeout in minutes.+:.+Timeout in minutes.+:.+",
		},
		{
			desc: "capabilities is empty",
			inputDefinitions: map[string]string{
				"Capabilities": "\nn",
			},
			expectedOutputRegex: "(?s)Pass all capabilities below automatically, ok.+Pass all capabilities below automatically, ok.+",
		},
		{
			desc: "invalid input for acknowledge capabilities",
			inputDefinitions: map[string]string{
				"Capabilities": fmt.Sprintf("invalid_input\n%s", create_stack.NO),
			},
			expectedOutputRegex: "(?s)Pass all capabilities below automatically, ok.+Pass all capabilities below automatically, ok.+",
		},
		{
			desc: "onFailure is empty",
			inputDefinitions: map[string]string{
				"OnFailure": "\n2",
			},
			expectedOutputRegex: "(?s)[On failure].+[On failure].+",
		},
		{
			desc: "invalid input for onFailure",
			inputDefinitions: map[string]string{
				"OnFailure": "invalid_input\n2",
			},
			expectedOutputRegex: "(?s)[On failure].+[On failure].+",
		},
		{
			desc: "enableTerminationProtection is empty",
			inputDefinitions: map[string]string{
				"EnableTerminationProtection": fmt.Sprintf("\n%s", create_stack.NO),
			},
			expectedOutputRegex: "Enable termination protection?.+Enable termination protection?.+",
		},
		{
			desc: "invalid input for enableTerminationProtection",
			inputDefinitions: map[string]string{
				"EnableTerminationProtection": fmt.Sprintf("invalid_input\n%s", create_stack.NO),
			},
			expectedOutputRegex: "Enable termination protection?.+Enable termination protection?.+",
		},
	}

	for _, tt := range invalid_input_cases {
		t.Run(fmt.Sprintf("invalid_input/%s", tt.desc), func(t *testing.T) {
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

			ui := initializeUserInput()
			for k, v := range tt.inputDefinitions {
				ui.overwriteUserInput(k, v)
			}
			cmd.SetIn(ui.makeUserInputBuffer())
			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			_, _ = create_stack.ExecCreateStack(cmd, []string{})

			expectedOutput := compileRegex(t, tt.expectedOutputRegex)
			actualOutput := readOutput(t, b)
			assert.Regexp(t, expectedOutput, actualOutput)
		})
	}

	red_cases := []struct {
		desc             string
		inputDefinitions map[string]string
		expected_error   error
	}{
		{
			desc: "not found error for local template file",
			inputDefinitions: map[string]string{
				"FilePath": "no_such_file",
			},
			expected_error: errors.New("There was an error processing the template: open no_such_file: no such file or directory"),
		},
		{
			desc: "parse error for local template file",
			inputDefinitions: map[string]string{
				"FilePath": fmt.Sprintf("\n%s", invalidSampleLocalTemplateFilePath),
			},
			expected_error: errors.New("There was an error processing the template: json: cannot unmarshal string into Go value of type cloudformation.Template"),
		},
		{
			desc: "refuse to acknowledge to pass capabilities",
			inputDefinitions: map[string]string{
				"Capabilities": create_stack.NO,
			},
			expected_error: errors.New("operation cancelled."),
		},
	}

	for _, tt := range red_cases {
		t.Run(fmt.Sprintf("error/%s", tt.desc), func(t *testing.T) {
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

			ui := initializeUserInput()
			for k, v := range tt.inputDefinitions {
				ui.overwriteUserInput(k, v)
			}
			cmd.SetIn(ui.makeUserInputBuffer())
			b := bytes.NewBufferString("")
			cmd.SetOut(b)

			stackId, err := create_stack.ExecCreateStack(cmd, []string{})

			assert.Equal(t, "", stackId)
			assert.Equal(t, tt.expected_error, err)
			cm.AssertNumberOfCalls(t, "CreateStack", 0)
		})
	}

	/************************************
		AWS Error
	************************************/

	t.Run("authentication error for CreateStack", func(t *testing.T) {
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the CreateStack operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: cloudformation:CreateStack"
		cm := &create_stack.MockCfnClient{}
		cm.On("CreateStack", mock.AnythingOfType("*cloudformation.CreateStackInput")).Return(nil, awserr.New(errorCode, errorMsg, nil))
		sm := &create_stack.MockS3Client{}
		initMockClient(cm, sm)

		cmd := create_stack.NewCmd()

		ui := initializeUserInput()
		cmd.SetIn(ui.makeUserInputBuffer())

		stackId, err := create_stack.ExecCreateStack(cmd, []string{})

		assert.Empty(t, stackId)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "CreateStack", 1)
		sm.AssertNumberOfCalls(t, "GetObject", 0)
	})

	t.Run("authentication error for GetObject", func(t *testing.T) {
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the GetObject operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: s3:GetObject"
		cm := &create_stack.MockCfnClient{}
		sm := &create_stack.MockS3Client{}
		sm.On("GetObject", mock.AnythingOfType("*s3.GetObjectInput")).Return(nil, awserr.New(errorCode, errorMsg, nil))
		initMockClient(cm, sm)

		cmd := create_stack.NewCmd()

		ui := initializeUserInput()
		ui.TemplateInS3 = create_stack.YES
		cmd.SetIn(ui.makeUserInputBuffer())

		stackId, err := create_stack.ExecCreateStack(cmd, []string{})

		assert.Empty(t, stackId)
		assert.Equal(t, errors.New(fmt.Sprintf("There was an error processing the template: %s: %s", errorCode, errorMsg)), err)
		cm.AssertNumberOfCalls(t, "CreateStack", 0)
		sm.AssertNumberOfCalls(t, "GetObject", 1)
	})

	t.Run("could not read template file in S3", func(t *testing.T) {
		const errorCode = "AccessDeniedException"
		const errorMsg = "An error occurred (AccessDeniedException) when calling the GetObject operation: User: arn:aws:iam::xxxxx:user/xxxxx is not authorized to perform: s3:GetObject"
		cm := &create_stack.MockCfnClient{}
		cm.On("CreateStack", mock.AnythingOfType("*cloudformation.CreateStackInput"))
		sm := &create_stack.MockS3Client{}
		sm.On("GetObject", mock.AnythingOfType("*s3.GetObjectInput")).Return(nil, awserr.New(errorCode, errorMsg, nil))
		initMockClient(cm, sm)

		cmd := create_stack.NewCmd()

		ui := initializeUserInput()
		ui.TemplateInS3 = create_stack.YES
		cmd.SetIn(ui.makeUserInputBuffer())

		stackId, err := create_stack.ExecCreateStack(cmd, []string{})

		assert.Equal(t, "", stackId)
		assert.Equal(t, errors.New(fmt.Sprintf("There was an error processing the template: %s: %s", errorCode, errorMsg)), err)
		cm.AssertNumberOfCalls(t, "CreateStack", 0)
		sm.AssertNumberOfCalls(t, "GetObject", 1)
	})

	t.Run("could not parse template file in S3", func(t *testing.T) {
		cm := &create_stack.MockCfnClient{}
		cm.On("CreateStack", mock.AnythingOfType("*cloudformation.CreateStackInput"))
		sm := &create_stack.MockS3Client{}
		sm.On("GetObject", mock.AnythingOfType("*s3.GetObjectInput")).Return(
			&s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("invalid format")),
			},
			nil,
		)
		initMockClient(cm, sm)

		cmd := create_stack.NewCmd()

		ui := initializeUserInput()
		ui.TemplateInS3 = create_stack.YES
		cmd.SetIn(ui.makeUserInputBuffer())

		stackId, err := create_stack.ExecCreateStack(cmd, []string{})

		assert.Equal(t, "", stackId)
		assert.Equal(t, errors.New("There was an error processing the template: json: cannot unmarshal string into Go value of type cloudformation.Template"), err)
		cm.AssertNumberOfCalls(t, "CreateStack", 0)
		sm.AssertNumberOfCalls(t, "GetObject", 1)
	})

	t.Run("already exists exception for CreateStack", func(t *testing.T) {
		const errorCode = "AlreadyExistsException"
		const errorMsg = "Stack [sample-stack-name] already exists"
		cm := &create_stack.MockCfnClient{}
		cm.On("CreateStack", mock.AnythingOfType("*cloudformation.CreateStackInput")).Return(nil, awserr.New(errorCode, errorMsg, nil))
		sm := &create_stack.MockS3Client{}
		initMockClient(cm, sm)

		cmd := create_stack.NewCmd()

		ui := initializeUserInput()
		cmd.SetIn(ui.makeUserInputBuffer())

		stackId, err := create_stack.ExecCreateStack(cmd, []string{})

		assert.Empty(t, stackId)
		assert.Equal(t, errorCode, err.(awserr.Error).Code())
		assert.Equal(t, errorMsg, err.(awserr.Error).Message())
		cm.AssertNumberOfCalls(t, "CreateStack", 1)
		sm.AssertNumberOfCalls(t, "GetObject", 0)
	})
}
