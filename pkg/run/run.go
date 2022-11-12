package run

import (
	"fmt"
	"html/template"
	"os"

	"github.com/fatih/color"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/manifoldco/promptui"
)

func Select(file string) get.Action {
	actions := config.GetConfig(file).Actions

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

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
	}
	chosenAct := actions[i]
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("You Chose: %s.\n", cyan(chosenAct.Name))
	PromptArguments(&chosenAct)
	get.Confirm(chosenAct)
	fmt.Println(chosenAct)
	return chosenAct
}

func PromptArguments(act *get.Action) {

	argsWithValues := act.Args
	// Each template displays the data received from the prompt with some formatting.
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ .Name }} ({{ .Value | faint }}): ",
		Valid:   "{{ .Name | green }} ({{ .Value | faint }}): ",
		Invalid: "{{ .Name | red }} ({{ .Value | faint }}): ",
		Success: "{{ .Name | bold }} ({{ .Value | faint }}): ",
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
	tmpl := `--------- Running "{{.Name}}" with: ----------
{{range .Args}}
{{ .Name }}: {{ .Value }}
{{end}}
	`

	t := template.Must(template.New("tmpl").Parse(tmpl))

	err := t.Execute(os.Stdout, act)
	if err != nil {
		fmt.Println("executing template:", err)
	}
}
