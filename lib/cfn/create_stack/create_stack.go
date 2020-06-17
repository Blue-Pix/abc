package create_stack

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
	notificationArns            string
	roleArn                     string
	onFailure                   string
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
	scanner := bufio.NewScanner(cmd.InOrStdin())
	scanner.Split(bufio.ScanLines)

	stackName = scan(cmd, scanner, "Stack name: ")
	templateInS3 = askBool(cmd, scanner, "Template in S3? (y or n): ")
	askTemplateFile(cmd, scanner)

	if err := askParameters(cmd, scanner); err != nil {
		return err
	}
	askTimeoutInMinutes(cmd, scanner)
	notificationArns = optionalScan(cmd, scanner, "Notification arns (comma separated): ")
	if !askBool(cmd, scanner, capabilitiesMessage) {
		return errors.New("operation cancelled.")
	}
	roleArn = optionalScan(cmd, scanner, "Role arn: ")
	askOnFailure(cmd, scanner)
	askTags(cmd, scanner)
	clientRequestToken = optionalScan(cmd, scanner, "Client request token: ")
	enableTerminationProtection = askBool(cmd, scanner, "Enable termination protection? (y or n): ")
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
	keys := util.SortStringMap(parameters)
	for _, k := range keys {
		p = append(p, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(parameters[k]),
		})
	}
	return p
}

func buildNotificationArnsInput() []*string {
	p := []*string{}
	arns := strings.Split(notificationArns, ",")
	for _, s := range arns {
		if s != "" {
			p = append(p, aws.String(s))
		}
	}
	return p
}

func buildOnFailureInput() string {
	switch onFailure {
	case "1":
		return "DO_NOTHING"
	case "2":
		return "ROLLBACK"
	case "3":
		return "DELETE"
	}
	return ""
}

func buildTagsInput() []*cloudformation.Tag {
	p := []*cloudformation.Tag{}
	keys := util.SortStringMap(tags)
	for _, k := range keys {
		p = append(p, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(tags[k]),
		})
	}
	return p
}

func createStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	return CfnClient.CreateStack(input)
}

func scan(cmd *cobra.Command, scanner *bufio.Scanner, msg string) string {
	cmd.Print(msg)
	for scanner.Scan() {
		input := scanner.Text()
		if input != "" {
			return input
		}
		cmd.Print(msg)
	}
	return ""
}

func optionalScan(cmd *cobra.Command, scanner *bufio.Scanner, msg string) string {
	cmd.Print(msg)
	scanner.Scan()
	return scanner.Text()
}

func askTemplateFile(cmd *cobra.Command, scanner *bufio.Scanner) {
	if templateInS3 {
		bucketName = scan(cmd, scanner, "S3 Bucket name: ")
		bucketKey = scan(cmd, scanner, "S3 Bucket key: ")
		bucketRegion = scan(cmd, scanner, "S3 Bucket region: ")
	} else {
		filePath = scan(cmd, scanner, "File path: ")
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

func askParameters(cmd *cobra.Command, scanner *bufio.Scanner) error {
	template, err := parseTemplate()
	if err != nil {
		return errors.New(fmt.Sprintf("There was an error processing the template: %s", err))
	}
	if len(template.Parameters) > 0 {
		cmd.Println("Parameters: ")
		keys := util.SortGeneralMap(template.Parameters)
		for _, k := range keys {
			v := template.Parameters[k]
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
			input := optionalScan(cmd, scanner, fmt.Sprint(" ", fmt.Sprintf("%s%s%s: ", k, desc, defaultValueMsg)))
			if input != "" {
				parameters[k] = input
			} else if defaultValue != "" {
				parameters[k] = defaultValue
			}
		}
	}
	return nil
}

func askTimeoutInMinutes(cmd *cobra.Command, scanner *bufio.Scanner) {
	const defaultValue = 60
	for {
		input := optionalScan(cmd, scanner, fmt.Sprintf("Timeout in minutes (default %d): ", defaultValue))
		if input == "" {
			timeoutInMinutes = defaultValue
			break
		} else {
			num, err := strconv.Atoi(input)
			if err == nil {
				timeoutInMinutes = int64(num)
				break
			}
		}
	}
}

func askOnFailure(cmd *cobra.Command, scanner *bufio.Scanner) {
	const msg = `
[On failure]
1. DO_NOTHING
2. ROLLBACK
3. DELETE
(type number): `
	cmd.Print(msg)
	for scanner.Scan() {
		input := scanner.Text()
		if input == "1" || input == "2" || input == "3" {
			onFailure = input
			break
		}
		cmd.Print(msg)
	}
}

func askTags(cmd *cobra.Command, scanner *bufio.Scanner) {
	cmd.Print("Tags (Key=Value comma separated): ")
	for scanner.Scan() {
		input := scanner.Text()
		if input == "" {
			break
		}
		valid := true
		for _, tag := range strings.Split(input, ",") {
			arr := strings.Split(tag, "=")
			if len(arr) != 2 {
				cmd.Println("[ERROR] invalid format.")
				valid = false
			}
			tags[arr[0]] = arr[1]
		}
		if valid {
			break
		}
	}
}

func askBool(cmd *cobra.Command, scanner *bufio.Scanner, msg string) bool {
	input := scan(cmd, scanner, msg)
	if input == "y" {
		return true
	} else if input == "n" {
		return false
	}
	return false
}
