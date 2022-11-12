package list

import (
	"fmt"
	"html/template"
	"os"

	"github.com/jatalocks/opsilon/internal/config"
)

func List(file string) {
	actions := config.GetConfig(file).Actions

	tmpl := `{{range .}}--------- {{.Name}} ----------
{{.Help}}
{{range .Args}}
- {{.Name}}{{end}}
{{end}}`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	err := t.Execute(os.Stdout, actions)
	if err != nil {
		fmt.Println("executing template:", err)
	}

}
