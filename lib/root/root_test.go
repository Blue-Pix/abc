package root

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/Blue-Pix/abc/lib/ami"
	"github.com/Blue-Pix/abc/lib/cfn"
	"github.com/Blue-Pix/abc/lib/cfn/unusedExports"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func prepareAmiCmd(args []string) *cobra.Command {
	cmd := NewCmd()
	cmd.SetArgs(args)
	amiCmd := ami.NewCmd()
	cmd.AddCommand(amiCmd)
	return cmd
}

func prepareCfnCmd(args []string) *cobra.Command {
	cmd := NewCmd()
	cmd.SetArgs(args)
	cfnCmd := cfn.NewCmd()
	unusedExportsCmd := unusedExports.NewCmd()
	cfnCmd.AddCommand(unusedExportsCmd)
	cmd.AddCommand(cfnCmd)
	return cmd
}

func TestExecute(t *testing.T) {
	t.Run("ami", func(t *testing.T) {
		t.Run("query by --version", func(t *testing.T) {
			args := []string{"ami", "--version", "2"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"version\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by -v", func(t *testing.T) {
			args := []string{"ami", "-v", "1"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"version\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by --virtualization-type", func(t *testing.T) {
			args := []string{"ami", "--virtualization-type", "hvm"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"virtualization_type\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by -V", func(t *testing.T) {
			args := []string{"ami", "-V", "pv"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"virtualization_type\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by --arch", func(t *testing.T) {
			args := []string{"ami", "--arch", "x86_64"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"arch\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by -a", func(t *testing.T) {
			args := []string{"ami", "-a", "arm64"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"arch\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by --storage", func(t *testing.T) {
			args := []string{"ami", "--storage", "gp2"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"storage\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by -s", func(t *testing.T) {
			args := []string{"ami", "-s", "ebs"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"storage\":\"([^\"]+)\"")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by --minimal", func(t *testing.T) {
			args := []string{"ami", "--minimal", "true"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"minimal\":([^\",]+),")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query by -m", func(t *testing.T) {
			args := []string{"ami", "-m", "false"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"minimal\":([^\",]+),")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
			}
		})

		t.Run("query with full options", func(t *testing.T) {
			args := []string{"ami", "-v", "2", "-V", "hvm", "-a", "x86_64", "-s", "gp2", "-m", "false"}
			cmd := prepareAmiCmd(args)
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.Execute()
			out, err := ioutil.ReadAll(b)
			if err != nil {
				t.Fatal(err)
			}
			r := regexp.MustCompile("\"version\":\"([^\"]+)\",\"virtualization_type\":\"([^\"]+)\",\"arch\":\"([^\"]+)\",\"storage\":\"([^\"]+)\",\"minimal\":([^\",]+),")
			list := r.FindAllStringSubmatch(string(out), -1)
			if len(list) == 0 {
				t.Fatal("there is no much result")
			}
			for _, l := range list {
				if l[1] != args[2] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[2], l[1]))
				}
				if l[2] != args[4] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[4], l[2]))
				}
				if l[3] != args[6] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[6], l[3]))
				}
				if l[4] != args[8] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[8], l[4]))
				}
				if l[5] != args[10] {
					t.Fatal(fmt.Sprintf("expected: %s, actual: %s", args[10], l[5]))
				}
			}
		})
	})

	t.Run("cfn", func(t *testing.T) {
		t.Run("unusedExports", func(t *testing.T) {
			t.Run("", func(t *testing.T) {
				args := []string{"cfn", "unused-exports"}
				cmd := prepareCfnCmd(args)
				b := bytes.NewBufferString("")
				cmd.SetOut(b)
				cmd.Execute()
				out, err := ioutil.ReadAll(b)
				if err != nil {
					t.Fatal(err)
				}
				list := strings.Split(string(out), "\n")
				assert.Regexp(t, regexp.MustCompile(`^\.+$`), list[0])
				assert.Equal(t, "name,exporting_stack", list[1])
				assert.Regexp(t, regexp.MustCompile(`^.+,.+$`), list[2])
			})
		})
	})
}
