package list

import (
	"html/template"
	"os"

	"github.com/jatalocks/opsilon/internal/logger"
	"github.com/jatalocks/opsilon/internal/utils"
)

func List(repoList []string) {
	workflows := utils.GetWorkflowsForRepo(repoList)

	tmpl := `{{range .}}
	--------- {{.ID}} ----------
Description: 
{{.Description}}
Input: {{range .Input}}
- {{.Name}} {{ if .Default}}({{.Default}}){{end}}{{end}}
{{end}}`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	err := t.Execute(os.Stdout, workflows)
	logger.HandleErr(err)

}
