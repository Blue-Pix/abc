package create_stack

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/awslabs/goformation"
	_cloudformation "github.com/awslabs/goformation/cloudformation"
	"github.com/spf13/cobra"
)

var CfnClient cloudformationiface.CloudFormationAPI
var S3Client s3iface.S3API

var (
	stackName    string
	templateInS3 bool
	filePath     string
	bucketName   string
	bucketRegion string
	bucketKey    string
	parameters   map[string]string
	// disableRollback             bool
	timeoutInMinutes            int64
	notificationArns            []string
	roleArn                     string
	onFailure                   int
	tags                        map[string]string
	clientRequestToken          string
	enableTerminationProtection bool
)

const capabilitiesMessage = `
[Confirmation]
Pass all capabilities below automatically, ok?
- CAPABILITY_IAM
- CAPABILITY_NAMED_IAM
- CAPABILITY_AUTO_EXPAND 
(y or n): `

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-stack",
		Short: "Create stack interactively.",
		Long: `
[abc cfn create-stack]
This command create CloudFormation's stack in interactive mode.
Following options are not supported.
- --disable-rollback (use --on-failure instead)
- --rollback-configuration
- --resource-types
- --stack-policy-body
- --stack-policy-url

Internally it uses aws cloudformation api.
Please configure your aws credentials with following policies.
- cloudformation:CreateStack`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	parameters = make(map[string]string)
	tags = make(map[string]string)
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	stackId, err := ExecCreateStack(cmd, args)
	if err != nil {
		return err
	}
	cmd.Printf("StackId: %sPerform create-stack is in progress asynchronously.\nPlease check creation status by yourself.\n", stackId)
	return nil
}

func ExecCreateStack(cmd *cobra.Command, args []string) (string, error) {
	initClient(cmd)
	if err := ask(cmd); err != nil {
		return "", err
	}
	input, err := BuildCreateStackInput()
	if err != nil {
		return "", err
	}
	output, err := createStack(input)
	if err != nil {
		return "", err
	}
	return aws.StringValue(output.StackId), nil
}

func ask(cmd *cobra.Command) error {
	askStackName(cmd)
	templateInS3 = askBool(cmd, "Template in S3? (y or n): ")
	askTemplateFile(cmd)
	if err := askParameters(cmd); err != nil {
		return err
	}
	// disableRollback = askBool(cmd, "Disable rollback?(default false) (y or n): ")
	askTimeoutInMinutes(cmd)
	askNotificationArns(cmd)
	if !askBool(cmd, capabilitiesMessage) {
		return errors.New("operation cancelled.")
	}
	askRoleArn(cmd)
	askOnFailure(cmd)
	askTags(cmd)
	askClientRequestToken(cmd)
	enableTerminationProtection = askBool(cmd, "Enable termination protection? (y or n): ")
	return nil
}

func initClient(cmd *cobra.Command) {
	profile, _ := cmd.Flags().GetString("profile")
	region, _ := cmd.Flags().GetString("region")
	if CfnClient == nil {
		sess := util.CreateSession(profile, region)
		CfnClient = cloudformation.New(sess)
	}
	if S3Client == nil {
		sess := util.CreateSession(profile, region)
		S3Client = s3.New(sess)
	}
}

func BuildCreateStackInput() (*cloudformation.CreateStackInput, error) {
	input := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
		// DisableRollback:  aws.Bool(disableRollback),
		TimeoutInMinutes: aws.Int64(timeoutInMinutes),
		Capabilities: []*string{
			aws.String("CAPABILITY_IAM"),
			aws.String("CAPABILITY_NAMED_IAM"),
			aws.String("CAPABILITY_AUTO_EXPAND"),
		},
		EnableTerminationProtection: aws.Bool(enableTerminationProtection),
	}

	if err := buildTemplateInput(input); err != nil {
		return nil, err
	}

	input.SetParameters(buildParametersInput())
	input.SetNotificationARNs(buildNotificationArnsInput())

	if roleArn != "" {
		input.SetRoleARN(roleArn)
	}

	input.SetOnFailure(buildOnFailureInput())
	input.SetTags(buildTagsInput())

	if clientRequestToken != "" {
		input.SetClientRequestToken(clientRequestToken)
	}

	return input, nil
}

func buildTemplateInput(input *cloudformation.CreateStackInput) error {
	if templateInS3 {
		input.SetTemplateURL(fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", bucketName, bucketRegion, bucketKey))
	} else {
		body, err := readTemplateBodyFromLocal()
		if err != nil {
			return err
		}
		input.SetTemplateBody(body)
	}
	return nil
}

func readTemplateBodyFromLocal() (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readTemplateBodyFromS3() ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(bucketKey),
	}
	fmt.Println(input)
	result, err := S3Client.GetObject(input)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	b, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func buildParametersInput() []*cloudformation.Parameter {
	p := []*cloudformation.Parameter{}
	for k, v := range parameters {
		p = append(p, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}
	return p
}

func buildNotificationArnsInput() []*string {
	p := []*string{}
	for _, s := range notificationArns {
		if s != "" {
			p = append(p, aws.String(s))
		}
	}
	return p
}

func buildOnFailureInput() string {
	switch onFailure {
	case 1:
		return "DO_NOTHING"
	case 2:
		return "ROLLBACK"
	case 3:
		return "DELETE"
	}
	return ""
}

func buildTagsInput() []*cloudformation.Tag {
	p := []*cloudformation.Tag{}
	for k, v := range tags {
		p = append(p, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return p
}

func createStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return CfnClient.CreateStack(input)
}

func askStackName(cmd *cobra.Command) {
	cmd.Print("Stack name: ")
	fmt.Scan(&stackName)
}

func askTemplateFile(cmd *cobra.Command) {
	if templateInS3 {
		cmd.Print("S3 Bucket name: ")
		fmt.Scan(&bucketName)
		cmd.Print("S3 Bucket region: ")
		fmt.Scan(&bucketRegion)
		cmd.Print("S3 Bucket key: ")
		fmt.Scan(&bucketKey)
	} else {
		cmd.Print("File path: ")
		fmt.Scan(&filePath)
	}
}

func parseTemplate() (*_cloudformation.Template, error) {
	if templateInS3 {
		b, err := readTemplateBodyFromS3()
		if err != nil {
			return nil, err
		}
		return goformation.ParseYAML(b)
	}
	return goformation.Open(filePath)
}

func askParameters(cmd *cobra.Command) error {
	template, err := parseTemplate()
	if err != nil {
		return errors.New(fmt.Sprintf("There was an error processing the template: %s", err))
	}
	if len(template.Parameters) > 0 {
		cmd.Println("Parameters: ")
		for k, v := range template.Parameters {
			var desc string
			var defaultValue string
			var defaultValueMsg string
			if v.(map[string]interface{})["Description"] != nil {
				desc = fmt.Sprint(" ", fmt.Sprintf("(%s)", v.(map[string]interface{})["Description"]))
			}
			if v.(map[string]interface{})["Default"] != nil {
				defaultValueMsg = fmt.Sprint(" ", fmt.Sprintf("[%s]", v.(map[string]interface{})["Default"]))
				defaultValue = v.(map[string]interface{})["Default"].(string)
			}
			cmd.Print(" ", fmt.Sprintf("%s%s%s: ", k, desc, defaultValueMsg))
			var input string
			fmt.Scanln(&input)
			if input != "" {
				parameters[k] = input
			} else if defaultValue != "" {
				parameters[k] = defaultValue
			}
		}
	}
	return nil
}

func askTimeoutInMinutes(cmd *cobra.Command) {
	const defaultValue = 60
	cmd.Printf("Timeout in minutes (default %d): ", defaultValue)
	fmt.Scanln(&timeoutInMinutes)
	if timeoutInMinutes == 0 {
		timeoutInMinutes = defaultValue
	}
}

func askNotificationArns(cmd *cobra.Command) {
	var input string
	cmd.Print("Notification arns (comma separated): ")
	fmt.Scanln(&input)
	input = strings.Trim(input, "\n")
	notificationArns = strings.Split(input, ",")
}

func askRoleArn(cmd *cobra.Command) {
	cmd.Print("Role arn: ")
	fmt.Scanln(&roleArn)
}

func askOnFailure(cmd *cobra.Command) {
	cmd.Print(`
[On failure]
1. DO_NOTHING
2. ROLLBACK
3. DELETE
type number: `)
	fmt.Scanln(&onFailure)
}

func askTags(cmd *cobra.Command) error {
	var input string
	cmd.Print("Tags (Key=Value comma separated): ")
	fmt.Scanln(&input)
	for _, tag := range strings.Split(input, ",") {
		arr := strings.Split(tag, "=")
		if len(arr) != 2 {
			return errors.New("invalid format.")
		}
		tags[arr[0]] = arr[1]
	}
	return nil
}

func askClientRequestToken(cmd *cobra.Command) {
	cmd.Print("Client request token: ")
	fmt.Scanln(&clientRequestToken)
}

func askBool(cmd *cobra.Command, msg string) bool {
	for {
		var input string
		cmd.Print(msg)
		fmt.Scan(&input)
		if input == "y" {
			return true
		} else if input == "n" {
			return false
		}
	}
}
