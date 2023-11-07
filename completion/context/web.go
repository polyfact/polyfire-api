package context

import (
	"text/template"

	"github.com/polyfire/api/web_request"
)

var promptWebTemplate = template.Must(
	template.New("web_context").Parse(`Some Content found on the web relevant to your task:
==========
{{range .Data}}{{.}}
{{end}}`),
)

type WebContext = TemplateContext

func GetWebContext(task string) (*WebContext, error) {
	res, err := webrequest.WebRequest(task)
	if err != nil || len(res) == 0 {
		return nil, err
	}

	return GetTemplateContext(res, *promptWebTemplate)
}
