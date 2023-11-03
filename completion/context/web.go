package context

import (
	"fmt"
	"text/template"

	"github.com/polyfire/api/web_request"
)

var PROMPT_WEB_TEMPLATE = template.Must(
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

	fmt.Println("RESULT WEB:", res)
	return GetTemplateContext(res, *PROMPT_WEB_TEMPLATE)
}
