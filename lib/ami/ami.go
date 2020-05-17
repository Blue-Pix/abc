package ami

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/Blue-Pix/abc/lib/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"
)

const PATH = "/aws/service/ami-amazon-linux-latest"

var (
	version            string
	virtualizationType string
	arch               string
	storage            string
	minimal            string
)

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
	str, err := Run(cmd, args)
	if err != nil {
		return err
	}
	cmd.Println(str)
	return nil
}

func Run(cmd *cobra.Command, args []string) (string, error) {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return "", err
	}
	sess := util.CreateSession(profile)

	amis, err := getAMIList(sess)
	if err != nil {
		return "", err
	}

	if version != "" {
		amis = filterByVersion(version, amis)
	}
	if virtualizationType != "" {
		amis = filterByVirtualizationType(virtualizationType, amis)
	}
	if arch != "" {
		amis = filterByArch(arch, amis)
	}
	if storage != "" {
		amis = filterByStorage(storage, amis)
	}
	if minimal != "" {
		amis = filterByMinimal(minimal, amis)
	}

	str, err := toJSON(amis)
	if err != nil {
		return "", nil
	}
	return str, nil
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

func getParametersByPath(sess *session.Session, token *string, path string) (*ssm.GetParametersByPathOutput, error) {
	service := ssm.New(sess)
	params := &ssm.GetParametersByPathInput{
		NextToken: token,
		Path:      aws.String(path),
		Recursive: aws.Bool(true),
	}
	return service.GetParametersByPath(params)
}

func toAMI(parameter *ssm.Parameter) AMI {
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

func getAMIList(sess *session.Session) ([]AMI, error) {
	var parameters []*ssm.Parameter
	var token *string = nil
	var amis []AMI

	for {
		resp, err := getParametersByPath(sess, token, PATH)
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
		amis = append(amis, toAMI(parameter))
	}
	return amis, nil
}

func filterByVersion(version string, amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		if version == ami.Version {
			newAmis = append(newAmis, ami)
		}
	}
	return newAmis
}

func filterByVirtualizationType(virtualizationType string, amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		if virtualizationType == ami.VirtualizationType {
			newAmis = append(newAmis, ami)
		}
	}
	return newAmis
}

func filterByArch(arch string, amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		if arch == ami.Arch {
			newAmis = append(newAmis, ami)
		}
	}
	return newAmis
}

func filterByStorage(storage string, amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		if storage == ami.Storage {
			newAmis = append(newAmis, ami)
		}
	}
	return newAmis
}

func filterByMinimal(minimal string, amis []AMI) []AMI {
	var newAmis []AMI
	for _, ami := range amis {
		m, _ := strconv.ParseBool(minimal)
		if m == ami.Minimal {
			newAmis = append(newAmis, ami)
		}
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
