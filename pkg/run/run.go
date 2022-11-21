package run

import (
	"errors"
	"fmt"
	"html/template"
	"os"

	"github.com/fatih/color"
	"github.com/jatalocks/opsilon/internal/concurrency"
	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/get"
	"github.com/jatalocks/opsilon/internal/internaltypes"
	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/jatalocks/opsilon/internal/utils"
	"github.com/manifoldco/promptui"
	"golang.org/x/exp/slices"
)

func ValidateWorkflowArgs(repoName string, workflowName string, args map[string]string) ([]string, internaltypes.Workflow) {
	missing := []string{}
	wArgs := internaltypes.WorkflowArgument{Repo: repoName, Workflow: workflowName, Args: args}
	repoList := config.GetRepoList()
	if !slices.Contains(repoList, wArgs.Repo) {
		logger.Error(fmt.Sprint("Repo ", repoName, "is not in repository list - To view all, run opsilon repo list."))
		missing = append(missing, "repo")
	}
	workflows, err := get.GetWorkflowsForRepo([]string{repoName})
	logger.HandleErr(err)
	wFound := false
	chosenAct := internaltypes.Workflow{}
	for _, v := range workflows {
		if v.ID == workflowName {
			wFound = true
			chosenAct = v
		}
	}
	if !wFound {
		logger.Error(fmt.Sprint("Worklow ", workflowName, "not found in repository", repoName, " - To view all, run opsilon list."))
		missing = append(missing, "workflow")
	}
	err = InputArgsIntoWorklow(args, &chosenAct)
	if err != nil {
		missing = append(missing, "args")
	}
	return missing, chosenAct
}

func Select(repoName string, workflowName string, args map[string]string, confirm bool) {
	missing, chosenAct := ValidateWorkflowArgs(repoName, workflowName, args)
	fmt.Println("Missing", missing)
	chosenRepo := repoName
	if slices.Contains(missing, "repo") {
		repoList := config.GetRepoList()
		promptRepo := &promptui.Select{
			Label: "Select Repo",
			Items: repoList,
		}
		iR, _, err := promptRepo.Run()
		logger.HandleErr(err)
		chosenRepo = repoList[iR]
	}
	if slices.Contains(missing, "workflow") {
		workflows, err := get.GetWorkflowsForRepo([]string{chosenRepo})
		logger.HandleErr(err)
		chosenAct = internaltypes.Workflow{}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\u25B6\uFE0F {{ .ID | cyan }} ({{ .Description | green }})",
			Inactive: "  {{ .ID | cyan }} ({{ .Description | yellow }})",
			Selected: "\u25B6\uFE0F {{ .ID | cyan }}",
		}

		prompt := &promptui.Select{
			Label:     "Select Workflow",
			Items:     workflows,
			Templates: templates,
		}

		i, _, err := prompt.Run()

		logger.HandleErr(err)
		chosenAct = workflows[i]
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("You Chose: %s\n", cyan(chosenAct.ID))
	if slices.Contains(missing, "args") || slices.Contains(missing, "workflow") || slices.Contains(missing, "repo") {
		PromptArguments(&chosenAct)
	}
	if !confirm {
		confirm, _ = utils.Confirm(chosenAct)
	}
	if confirm {
		concurrency.ToGraph(chosenAct)
	} else {
		fmt.Println("Run Canceled")
	}
}

func InputArgsIntoWorklow(m map[string]string, act *internaltypes.Workflow) error {
	argsWithValues := act.Input
	for i, input := range argsWithValues {
		if val, ok := m[input.Name]; ok {
			argsWithValues[i].Default = val
		} else {
			if !input.Optional {
				logger.Error("Input", input.Name, "is mandatory but none was provided.")
				return errors.New("is mandatory but none was provided.")
			}
		}
	}
	return nil
}

func PromptArguments(act *internaltypes.Workflow) {
	argsWithValues := act.Input
	// Each template displays the data received from the prompt with some formatting.
	templates := &promptui.PromptTemplates{
		Prompt:  "{{ .Name }} ({{ .Default | faint }}): ",
		Valid:   "{{ .Name | green }} ({{ .Default | faint }}): ",
		Invalid: "{{ .Name | red }} ({{ .Default | faint }}): ",
		Success: "{{ .Name | bold }} ({{ .Default | faint }}): ",
	}

	for i, v := range argsWithValues {
		// The validate function follows the required validator signature.
		validate := func(input string) error {
			if input == "" && !v.Optional && v.Default == "" {
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
			result = v.Default
		}

		// The result of the prompt, if valid, is displayed in a formatted message.
		argsWithValues[i].Default = result
		fmt.Printf("%s\n", result)
	}
	tmpl := `--------- Running "{{.ID}}" with: ----------
{{range .Input}}
{{ .Name }}: {{ .Default }}
{{end}}
	`

	t := template.Must(template.New("tmpl").Parse(tmpl))

	err := t.Execute(os.Stdout, act)

	logger.HandleErr(err)
}
