package run

import (
	"fmt"

	"github.com/jatalocks/opsilon/internal/get"
	"github.com/manifoldco/promptui"
)

func Select(file string) string {
	actions := get.Actions(get.List(file))

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\u25B6\uFE0F {{ .Name | cyan }} ({{ .Help | green }})",
		Inactive: "  {{ .Name | cyan }} ({{ .Help | yellow }})",
		Selected: "\u25B6\uFE0F {{ .Name | cyan | cyan }}",
		Details: `
--------- Action ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Description:" | faint }}	{{ .Help }}
{{ "Arguments:" | faint }}	{{block "list" .Args}}{{"\n"}}{{range .}}{{println "-" .Name}}{{end}}{{end}}`,
	}

	prompt := promptui.Select{
		Label:     "Select Action",
		Items:     actions,
		Templates: templates,
	}

	i, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}

	fmt.Printf("You Chose: %s\n", actions[i].Name)
	PromptArguments(&actions[i].Args)
	return result
}

func PromptArguments(act *[]get.Argument) {

	argsWithValues := *act
	// Each template displays the data received from the prompt with some formatting.
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ .Type | faint }} {{ .Name }} ({{ .Value | faint }}): ",
		Valid:   "{{ .Type | faint }} {{ .Name | green }} ({{ .Value | faint }}): ",
		Invalid: "{{ .Type | faint }} {{ .Name | red }} ({{ .Value | faint }}): ",
		Success: "{{ .Type | faint }} {{ .Name | bold }} ({{ .Value | faint }}): ",
	}

	for i, v := range argsWithValues {
		// The validate function follows the required validator signature.
		validate := func(input string) error {
			if input == "" && !v.Optional && v.Value == "" {
				return fmt.Errorf("This argument is mandatory")
			}
			return nil
		}

		prompt := promptui.Prompt{
			Label:     v,
			Templates: templates,
			Validate:  validate,
		}

		result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if result == "" {
			result = v.Value
		}

		// The result of the prompt, if valid, is displayed in a formatted message.
		argsWithValues[i].Value = result
		fmt.Printf("%s\n", result)
	}
}

func Error(s string) {
	panic("unimplemented")
}
