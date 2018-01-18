package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type (
	server struct {
		mutex    sync.RWMutex
		handlers []handler
		requests []*request
	}

	handler func(w http.ResponseWriter, r *request) (bool, bool, error)

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
	}
}

//
// Handlers

func (s *server) apiHandler(w http.ResponseWriter, r *http.Request) {
	req, err := convertRequest(r)
	if err != nil {
		writeError(w, "Failed to convert request (%s)\n", err.Error())
		return
	}

	s.mutex.Lock()
	s.requests = append(s.requests, req)
	s.mutex.Unlock()

	for _, handler := range s.handlers {
		if matched, headersWritten, err := handler(w, req); matched {
			if err != nil {
				if headersWritten {
					w = nil
				}

				writeError(w, "Failed to invoke handler (%s)\n", err.Error())
			}

			return
		}
	}

	fmt.Printf("Did not find a matching handler for request\n")
	w.WriteHeader(http.StatusNotFound)
}

func (s *server) clearHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	s.handlers = s.handlers[:0]
	s.mutex.Unlock()
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {
	handler, err := makeHandler(r.Body)
	if err != nil {
		writeError(w, "Failed to make handler (%s)\n", err.Error())
		return
	}

	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	s.mutex.Unlock()
}

func (s *server) gatherHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.Marshal(s.requests)
	if err != nil {
		writeError(w, "Failed to serialize request list (%s)\n", err.Error())
		return
	}

	if r.URL.Query().Get("clear") != "" {
		s.requests = s.requests[:0]
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
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

	handler := func(w http.ResponseWriter, r *request) (bool, bool, error) {
		if !expectation.Matches(r) {
			return false, false, nil
		}

		headersWritten, err := template.Write(w, r)
		return true, headersWritten, err
	}

	return handler, nil
}

func writeError(w http.ResponseWriter, format string, args ...interface{}) {
	fmt.Printf(format, args...)

	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
