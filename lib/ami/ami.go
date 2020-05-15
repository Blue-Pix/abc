package ami

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/spf13/cobra"
)

const PATH = "/aws/service/ami-amazon-linux-latest"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ami",
		Short: "return latest amazon linux ami",
		Long: `[abc ami]
	This command returns latest amazon linux ami as json format.
		
	Internally it uses ssm get-parameters-by-path api. (https://docs.aws.amazon.com/ja_jp/systems-manager/latest/userguide/parameter-store-public-parameters.html)
	Please configure your aws credential which has required policy.

	By default, this returns serveral type of amis.
	You can query it with options below. `,
		Run: run,
	}
	cmd.Flags().StringP("version", "v", "", "os version(1 or 2)")
	cmd.Flags().StringP("virtualization-type", "V", "", "virtualization type(hvm or pv)")
	cmd.Flags().StringP("arch", "a", "", "cpu architecture(x86_64 or arm64)")
	cmd.Flags().StringP("storage", "s", "", "storage type(gp2, ebs or s3)")
	cmd.Flags().StringP("minimal", "m", "", "if minimal image or not(true or false)")
	return cmd
}

func run(cmd *cobra.Command, args []string) {
	str := Run(cmd, args)
	cmd.Println(str)
}

func Run(cmd *cobra.Command, args []string) string {
	amis := getAMIList()
	if version, err := cmd.Flags().GetString("version"); version != "" && err == nil {
		amis = filterByVersion(version, amis)
	}
	if virtualizationType, err := cmd.Flags().GetString("virtualization-type"); virtualizationType != "" && err == nil {
		amis = filterByVirtualizationType(virtualizationType, amis)
	}
	if arch, err := cmd.Flags().GetString("arch"); arch != "" && err == nil {
		amis = filterByArch(arch, amis)
	}
	if storage, err := cmd.Flags().GetString("storage"); storage != "" && err == nil {
		amis = filterByStorage(storage, amis)
	}
	if minimal, err := cmd.Flags().GetString("minimal"); minimal != "" && err == nil {
		amis = filterByMinimal(minimal, amis)
	}
	str := toJSON(amis)
	return str
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

func getAMIList() []AMI {
	sess := session.Must(session.NewSession())
	var parameters []*ssm.Parameter
	var token *string = nil

	for {
		resp, err := getParametersByPath(sess, token, PATH)
		if err != nil {
			panic(err)
		}
		parameters = append(parameters, resp.Parameters...)

		if resp.NextToken == nil {
			break
		}
		token = resp.NextToken
	}

	var amis []AMI
	for _, parameter := range parameters {
		amis = append(amis, toAMI(parameter))
	}
	return amis
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

func toJSON(amis []AMI) string {
	jsonBytes, err := json.Marshal(amis)
	if err != nil {
		panic(err)
		fmt.Println(err)
	}
	jsonStr := string(jsonBytes)
	return jsonStr
}
