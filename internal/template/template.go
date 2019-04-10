package template

import (
	"bytes"
	"fmt"
	"strconv"
	tmpl "text/template"

	"github.com/efritz/derision/internal/expectation"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
)

type (
	Template interface {
		Respond(r *request.Request, m *expectation.Match) (response.Response, error)
	}

	template struct {
		statusCode *tmpl.Template
		headers    map[string][]*tmpl.Template
		body       *tmpl.Template
	}
)

var ErrIllegalStatusCode = fmt.Errorf("illegal status code")

func (t *template) Respond(r *request.Request, m *expectation.Match) (response.Response, error) {
	args := map[string]interface{}{
		"Method":       r.Method,
		"Path":         r.Path,
		"Headers":      r.Headers,
		"Body":         r.Body,
		"MethodGroups": m.MethodGroups,
		"PathGroups":   m.PathGroups,
		"HeaderGroups": m.HeaderGroups,
		"BodyGroups":   m.BodyGroups,
	}

	body, err := applyTemplate(t.body, args)
	if err != nil {
		return nil, err
	}

	resp := response.Respond([]byte(body))

	statusCode, err := applyTemplate(t.statusCode, args)
	if err != nil {
		return nil, err
	}

	if statusCode != "" {
		val, err := strconv.Atoi(statusCode)
		if err != nil {
			return nil, ErrIllegalStatusCode
		}

		resp.SetStatusCode(val)
	}

	for header, values := range t.headers {
		for _, value := range values {
			val, err := applyTemplate(value, args)
			if err != nil {
				return nil, err
			}

			resp.AddHeader(header, val)
		}
	}

	return resp, nil
}

func applyTemplate(t *tmpl.Template, args map[string]interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	if err := t.Execute(buffer, args); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
