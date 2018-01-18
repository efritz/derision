package main

import (
	"fmt"
	"strconv"

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
	statusCode, err := t.getStatusCode()
	if err != nil {
		return nil, err
	}

	resp := response.Respond([]byte(t.Body)).SetStatusCode(statusCode)

	for header, values := range t.Headers {
		for _, value := range values {
			resp.Header.Add(header, value)
		}
	}

	return resp, nil
}

func (t *template) getStatusCode() (int, error) {
	if t.StatusCode == "" {
		return 200, nil
	}

	statusCode, err := strconv.Atoi(t.StatusCode)
	if err != nil {
		return 0, err
	}

	return statusCode, nil
}
