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
		Path    string              `json:"path"`
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

func (s *server) registerHandler(r *http.Request) *response.Response {
	handler, err := makeHandler(r.Body)
	if err != nil {
		if verr, ok := err.(*validationError); ok {
			resp := s.makeDetailedError(verr.errors, "Validation error")
			resp.SetStatusCode(http.StatusBadRequest)
			return resp
		}

		return s.makeError("Failed to make handler (%s)", err.Error())
	}

	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	s.mutex.Unlock()

	return response.Empty(http.StatusNoContent)
}

func (s *server) clearHandler(r *http.Request) *response.Response {
	s.mutex.Lock()
	s.handlers = s.handlers[:0]
	s.mutex.Unlock()

	return response.Empty(http.StatusNoContent)
}

func (s *server) requestsHandler(r *http.Request) *response.Response {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	requests := s.requests

	if r.URL.Query().Get("clear") != "" {
		s.requests = s.requests[:0]
	}

	return response.JSON(requests)
}

func (s *server) apiHandler(r *http.Request) *response.Response {
	req, err := convertRequest(r)
	if err != nil {
		return s.makeError("Failed to convert request (%s)", err.Error())
	}

	s.logRequest(req)

	for _, handler := range s.handlers {
		response, err := handler(req)
		if err != nil {
			return s.makeError("Failed to invoke handler (%s)", err.Error())
		}

		if response != nil {
			return response
		}
	}

	resp := s.makeError("No matching handler registered for request")
	resp.SetStatusCode(http.StatusNotFound)
	return resp
}

func (s *server) logRequest(req *request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Add this request to the log
	s.requests = append(s.requests, req)

	// Ensure we don't use unlimited memory if no
	// one is checking the control endpoints
	if len(s.requests) > requestLogCapacity {
		s.requests = s.requests[1:]
	}
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

	resp := response.JSON(payload)
	resp.SetStatusCode(http.StatusInternalServerError)
	return resp
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
		Path:    r.URL.Path,
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
		if match := expectation.Matches(r); match != nil {
			return template.Respond(r, match)
		}

		return nil, nil
	}

	return handler, nil
}
