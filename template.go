package main

import (
	"net/http"
	"strconv"
)

type (
	template struct {
		StatusCode string              `json:"status_code"`
		Headers    map[string][]string `json:"headers"`
		Body       string              `json:"body"`
	}
)

func (t *template) Write(w http.ResponseWriter, r *request) (bool, error) {
	for header, values := range t.Headers {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	statusCode, err := t.getStatusCode()
	if err != nil {
		return false, err
	}

	w.WriteHeader(statusCode)

	return true, writeAll(w, []byte(t.Body))
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
