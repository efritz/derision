package main

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	tmpl "text/template"

	"github.com/efritz/response"
)

type (
	template struct {
		StatusCode string              `json:"status_code"`
		Headers    map[string][]string `json:"headers"`
		Body       string              `json:"body"`
	}
)

func (t *template) Respond(r *request, m *match) (*response.Response, error) {
	args := map[string]interface{}{
		"Method":       r.Method,
		"Path":         r.Path,
		"Headers":      r.Headers,
		"Body":         r.Body,
		"MethodGroups": m.methodGroups,
		"PathGroups":   m.pathGroups,
		"HeaderGroups": m.headerGroups,
	}

	body, err := applyTemplate("body", t.Body, args)
	if err != nil {
		return nil, err
	}

	resp := response.Respond([]byte(body))

	statusCode, err := applyTemplate("statusCode", t.StatusCode, args)
	if err != nil {
		return nil, err
	}

	if statusCode != "" {
		val, err := strconv.Atoi(statusCode)
		if err != nil {
			return nil, err
		}

		resp.SetStatusCode(val)
	}

	for header, values := range t.Headers {
		for _, value := range values {
			val, err := applyTemplate(fmt.Sprintf("header-%s", header), value, args)
			if err != nil {
				return nil, err
			}

			resp.Header.Add(header, val)
		}
	}

	return resp, nil
}

//
// Helpers

func applyTemplate(name, text string, args map[string]interface{}) (string, error) {
	t, err := parseTemplate(name, text)
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)

	if err = t.Execute(buffer, args); err != nil {
		return "", err
	}

	return buffer.String(), err
}

var lock sync.Mutex

func parseTemplate(name, text string) (*tmpl.Template, error) {
	// Somehow, text/template is not thread-safe.
	lock.Lock()
	defer lock.Unlock()
	return tmpl.New(name).Parse(text)
}
