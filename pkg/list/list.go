package list

import (
	"html/template"
	"os"

	"github.com/jatalocks/opsilon/internal/config"
	"github.com/jatalocks/opsilon/internal/logger"
)

func List(file string) {
	actions := config.GetConfig(file)

	tmpl := `{{range .}}--------- {{.Name}} ----------
Description: 
	{{.Workflow.Description}}
Input: {{range .Workflow.Input}}
	- {{.Name}} {{ if .Value}}({{.Value}}){{end}}{{end}}
{{end}}`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	err := t.Execute(os.Stdout, actions)
	logger.HandleErr(err)

}
