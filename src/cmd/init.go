package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"regexp"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/i18n"
)

//go:embed resources/zerops_yml.tmpl
var zeropsYml string

var initTemplate = template.Must(template.New("zeropsYml").Parse(zeropsYml))

type templateData struct {
	Name string
}

var nameRegex = regexp.MustCompile("^[a-z][a-z0-9]{0,39}$")

func initCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init [name]",
		Short:   "initializes zerops project",
		Long:    `initializes zerops project by creating empty zerops.yml`,
		Example: "init app",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				fmt.Printf(i18n.PromptEnterZeropsServiceName + "\n\n")
				promptName, err := prompt[string](i18n.PromptName)
				if err != nil {
					return err
				}
				name = promptName
			}
			if !nameRegex.MatchString(name) {
				return errors.New(i18n.PromptInvalidHostname + " See: https://docs.zerops.io/documentation/export-import/project-service-export-import.html#yaml-specification")
			}
			file, err := os.Create("zerops.yml")
			if err != nil {
				return err
			}
			defer file.Close()
			return initTemplate.Execute(
				file,
				templateData{
					Name: name,
				},
			)
		},
		SilenceUsage: true,
	}

	return cmd
}

func prompt[T any](prompt string) (T, error) {
	fmt.Print("\t" + prompt + ": ")
	var t T
	n, err := fmt.Scan(&t)
	if err != nil {
		return t, err
	}
	if n == 0 {
		return t, errors.New(i18n.PromptInvalidInput)
	}
	return t, nil
}
