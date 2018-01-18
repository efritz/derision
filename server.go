package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/efritz/response"
)

type (
	server struct {
		handlers []handler
		requests []*request
		logger   *log.Logger
		mutex    sync.RWMutex
	}

	handler func(r *request) (*response.Response, error)

	request struct {
		Method  string              `json:"method"`
		URL     string              `json:"url"`
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	}
)

func newServer() *server {
	return &server{
		handlers: []handler{},
		requests: []*request{},
		logger:   log.New(os.Stdout, "[derision] ", 0),
	}
}

//
// Handlers

func (s *server) apiHandler(r *http.Request) *response.Response {
	req, err := convertRequest(r)
	if err != nil {
		return s.makeError("Failed to convert request (%s)", err.Error())
	}

	s.mutex.Lock()
	s.requests = append(s.requests, req)
	s.mutex.Unlock()

	for _, handler := range s.handlers {
		response, err := handler(req)
		if err != nil {
			return s.makeError("Failed to invoke handler (%s)", err.Error())
		}

		if response != nil {
			return response
		}
	}

	return s.makeError("No matching handler registered for request").SetStatusCode(http.StatusNotFound)
}

func (s *server) clearHandler(r *http.Request) *response.Response {
	s.mutex.Lock()
	s.handlers = s.handlers[:0]
	s.mutex.Unlock()

	return response.Empty(http.StatusNoContent)
}

func (s *server) registerHandler(r *http.Request) *response.Response {
	handler, err := makeHandler(r.Body)
	if err != nil {
		if v, ok := err.(*validationError); ok {
			errors := []string{}
			for _, err := range v.errors {
				errors = append(errors, fmt.Sprintf("%s: %s", err.Field(), err.Description()))
			}

			return s.makeDetailedError(errors, "Validation error").SetStatusCode(http.StatusBadRequest)
		}

		return s.makeError("Failed to make handler (%s)", err.Error())
	}

	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	s.mutex.Unlock()

	return response.Empty(http.StatusNoContent)
}

func (s *server) gatherHandler(r *http.Request) *response.Response {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requests := s.requests

	if r.URL.Query().Get("clear") != "" {
		s.requests = s.requests[:0]
	}

	return response.JSON(requests)
}

//
// Errors

func (s *server) makeError(format string, args ...interface{}) *response.Response {
	return s.makeDetailedError(nil, format, args...)
}

func (s *server) makeDetailedError(details []string, format string, args ...interface{}) *response.Response {
	s.logger.Printf(format, args...)

	payload := map[string]interface{}{
		"message": fmt.Sprintf(format, args...),
	}

	if len(details) > 0 {
		payload["details"] = details
	}

	return response.JSON(payload).SetStatusCode(http.StatusInternalServerError)
}

//
// Helpers

func convertRequest(r *http.Request) (*request, error) {
	data, err := readAll(r.Body)
	if err != nil {
		return nil, err
	}

	return &request{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: r.Header,
		Body:    string(data),
	}, nil
}

func makeHandler(r io.ReadCloser) (handler, error) {
	expectation, template, err := readAndValidate(r)
	if err != nil {
		return nil, err
	}

	handler := func(r *request) (*response.Response, error) {
		if !expectation.Matches(r) {
			return nil, nil
		}

		return template.Respond(r)
	}

	return handler, nil
}
