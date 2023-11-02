package completion

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/polyfire/api/web_request"
)


const PROMPT_WEB_TEMPLATE = `
Date: {{.Date}}
From this request: {{.Task}}
and using this website content: {{.Content}}
answer at the above request and don't include appendix information other than the initial request.
Don't be creative, just be factual.
Always answer with the same language as the request.
If you don't know the answer, just say so.
If website content is not enough, you can use your own knowledge.
Use All websites content to make your relevant answer.`

func webContext(task string, model *string) (string, error) {
	res, err := webrequest.WebRequest(task, model)
	if err != nil {
		return "", err
	}

	data := struct {
		Task    string
		Content string
		Date    string
	}{
		Task:    task,
		Content: res,
		Date:    time.Now().Format("2006-01-02"),
	}

	var tpl bytes.Buffer
	t := template.Must(template.New("prompt").Parse(PROMPT_WEB_TEMPLATE))

	if err := t.Execute(&tpl, data); err != nil {
		fmt.Println("Error executing template:", err)
		return "", err
	}

	return tpl.String(), nil
}
