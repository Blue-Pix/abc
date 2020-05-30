package ami

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/spf13/cobra"
)

const PATH = "/aws/service/ami-amazon-linux-latest"

// flag
var (
	version            string
	virtualizationType string
	arch               string
	storage            string
	minimal            string
)

// mockable
var Client ssmiface.SSMAPI

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ami",
		Short: "Returns latest amazon linux ami",
		Long: `
[abc ami]
This command returns latest amazon linux ami as json format.
	
Internally it uses ssm get-parameters-by-path api. (https://docs.aws.amazon.com/ja_jp/systems-manager/latest/userguide/parameter-store-public-parameters.html)
Please configure your aws credential which has required policy.

By default, this returns serveral type of amis.
You can query it with options below. `,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := run(cmd, args)
			return err
		},
	}
	cmd.Flags().StringVarP(&version, "version", "v", "", "os version(1 or 2)")
	cmd.Flags().StringVarP(&virtualizationType, "virtualization-type", "V", "", "virtualization type(hvm or pv)")
	cmd.Flags().StringVarP(&arch, "arch", "a", "", "cpu architecture(x86_64 or arm64)")
	cmd.Flags().StringVarP(&storage, "storage", "s", "", "storage type(gp2, ebs or s3)")
	cmd.Flags().StringVarP(&minimal, "minimal", "m", "", "if minimal image or not(true or false)")
	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	amis, err := FetchData(cmd, args)
	if err != nil {
		return err
	}
	str, err := toJSON(amis)
	if err != nil {
		return err
	}
	cmd.Println(str)
	return nil
}

func FetchData(cmd *cobra.Command, args []string) ([]AMI, error) {
	initClient(cmd)
	amis, err := getAMIList()
	if err != nil {
		return nil, err
	}
	amis = filter(amis)
	return amis, nil
}

func initClient(cmd *cobra.Command) {
	if Client == nil {
		profile, _ := cmd.Flags().GetString("profile")
		region, _ := cmd.Flags().GetString("region")
		sess := util.CreateSession(profile, region)
		Client = ssm.New(sess)
	}
}

type AMI struct {
	Os                 string `json:"os"`
	Version            string `json:"version"`
	VirtualizationType string `json:"virtualization_type"`
	Arch               string `json:"arch"`
	Storage            string `json:"storage"`
	Minimal            bool   `json:"minimal"`
	Id                 string `json:"id"`
	Arn                string `json:"arn"`
}

func getParametersByPath(token *string, path string) (*ssm.GetParametersByPathOutput, error) {
	params := &ssm.GetParametersByPathInput{
		NextToken: token,
		Path:      aws.String(path),
		Recursive: aws.Bool(true),
	}
	return Client.GetParametersByPath(params)
}

func getAMIList() ([]AMI, error) {
	var parameters []*ssm.Parameter
	var token *string = nil
	var amis []AMI

	for {
		resp, err := getParametersByPath(token, PATH)
		if err != nil {
			return amis, err
		}
		parameters = append(parameters, resp.Parameters...)

		if resp.NextToken == nil {
			break
		}
		token = resp.NextToken
	}

	for _, parameter := range parameters {
		amis = append(amis, ToAMI(parameter))
	}
	return amis, nil
}

func ToAMI(parameter *ssm.Parameter) AMI {
	r := regexp.MustCompile(`^` + PATH + `\/([^\d]+)(\d)?-ami-(minimal\-)?(.+)-(.+)-(.+)$`)
	list := r.FindAllStringSubmatch(aws.StringValue(parameter.Name), -1)
	ami := AMI{
		Os:                 list[0][1],
		Version:            "1",
		VirtualizationType: list[0][4],
		Arch:               list[0][5],
		Storage:            list[0][6],
		Id:                 aws.StringValue(parameter.Value),
		Arn:                aws.StringValue(parameter.ARN),
	}
	if list[0][2] != "" {
		ami.Version = list[0][2]
	}
	if list[0][3] != "" {
		ami.Minimal = true
	}
	return ami
}

func filter(amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		if version != "" && version != ami.Version {
			continue
		}
		if virtualizationType != "" && virtualizationType != ami.VirtualizationType {
			continue
		}
		if arch != "" && arch != ami.Arch {
			continue
		}
		if storage != "" && storage != ami.Storage {
			continue
		}
		if minimal != "" {
			m, _ := strconv.ParseBool(minimal)
			if m != ami.Minimal {
				continue
			}
		}
		newAmis = append(newAmis, ami)
	}
	return newAmis
}

func toJSON(amis []AMI) (string, error) {
	jsonBytes, err := json.Marshal(amis)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)
	return jsonStr, nil
}
