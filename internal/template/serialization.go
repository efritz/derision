package template

import (
	"encoding/json"
	"fmt"
	tmpl "text/template"
)

type jsonTemplate struct {
	StatusCode string              `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

func Unmarshal(payload []byte) (Template, error) {
	t := &jsonTemplate{}
	if err := json.Unmarshal(payload, &t); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload (%s)", err.Error())
	}

	statusCode, err := compile(t.StatusCode)
	if err != nil {
		return nil, fmt.Errorf("illegal status code template")
	}

	headers := map[string][]*tmpl.Template{}
	for name, values := range t.Headers {
		templates := []*tmpl.Template{}
		for _, value := range values {
			template, err := compile(value)
			if err != nil {
				return nil, fmt.Errorf("illegal header template")
			}

			templates = append(templates, template)
		}

		headers[name] = templates
	}

	body, err := compile(t.Body)
	if err != nil {
		return nil, fmt.Errorf("illegal body template")
	}

	return &template{
		statusCode: statusCode,
		headers:    headers,
		body:       body,
	}, nil
}

func compile(template string) (*tmpl.Template, error) {
	return tmpl.New("").Parse(template)
}
