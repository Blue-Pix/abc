package create_stack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/awslabs/goformation"
	"github.com/spf13/cobra"
)

var CfnClient cloudformationiface.CloudFormationAPI

var (
	stackName    string
	templateInS3 bool
	filePath     string
	bucketName   string
	bucketRegion string
	bucketKey    string
	parameters   map[string]string
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-stack",
		Short: "Create stack interactively.",
		Long: `
[abc cfn create-stack]
This command create CloudFormation's stack in interactive mode.

Internally it uses aws cloudformation api.
Please configure your aws credentials with following policies.
- cloudformation:CreateStack`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	parameters = make(map[string]string)
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if err := ExecCreateStack(cmd, args); err != nil {
		return err
	}
	// cmd.Println("Perform create-stack is in progress asynchronously.\nPlease check creation status by yourself.")
	return nil
}

func ExecCreateStack(cmd *cobra.Command, args []string) error {
	initClient(cmd)

	cmd.Print("Stack name: ")
	fmt.Scan(&stackName)
	for {
		var input string
		cmd.Print("Template in S3? (y or n): ")
		fmt.Scan(&input)
		if input == "y" {
			templateInS3 = true
			break
		} else if input == "n" {
			templateInS3 = false
			break
		}
	}
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

	// ToDo download from s3
	template, err := goformation.Open(filePath)
	if err != nil {
		log.Fatalf("There was an error processing the template: %s", err)
	}
	fmt.Println("Parameters: ")
	for k, v := range template.Parameters {
		var desc string
		var defaultValue string
		if v.(map[string]interface{})["Description"] != nil {
			desc = fmt.Sprint(" ", fmt.Sprintf("(%s)", v.(map[string]interface{})["Description"]))
		}
		if v.(map[string]interface{})["Default"] != nil {
			defaultValue = fmt.Sprint(" ", fmt.Sprintf("[%s]", v.(map[string]interface{})["Default"]))
		}
		cmd.Print(" ", fmt.Sprintf("%s%s%s: ", k, desc, defaultValue))
		var input string
		fmt.Scan(&input)
		parameters[k] = input
	}

	output, err := createStack()
	if err != nil {
		return err
	}
	cmd.Println(aws.StringValue(output.StackId))
	return nil
}

func initClient(cmd *cobra.Command) {
	profile, _ := cmd.Flags().GetString("profile")
	region, _ := cmd.Flags().GetString("region")
	if CfnClient == nil {
		sess := util.CreateSession(profile, region)
		CfnClient = cloudformation.New(sess)
	}
}

func createStack() (*cloudformation.CreateStackOutput, error) {
	params := &cloudformation.CreateStackInput{
		StackName: aws.String(stackName),
	}
	if templateInS3 {
		params.TemplateURL = aws.String(fmt.Sprintf("https://%s.s3-%s.amazonaws.com/%s", bucketName, bucketRegion, bucketKey))
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		params.TemplateBody = aws.String(string(b))
	}

	if len(parameters) > 0 {
		p := make([]*cloudformation.Parameter, len(parameters))
		for k, v := range parameters {
			p = append(p, &cloudformation.Parameter{
				ParameterKey:   aws.String(k),
				ParameterValue: aws.String(v),
			})
		}
		params.SetParameters(p)
	}

	return CfnClient.CreateStack(params)
}
