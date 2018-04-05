package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/efritz/response"
)

type (
	server struct {
		maxRequestLog int
		handlers      []handler
		requests      []*request
		logger        *log.Logger
		mutex         sync.RWMutex
	}

	handler func(r *request) (*response.Response, error)

	request struct {
		Method   string              `json:"method"`
		Path     string              `json:"path"`
		Headers  map[string][]string `json:"headers"`
		Body     string              `json:"body"`
		RawBody  string              `json:"raw_body"`
		Form     map[string][]string `json:"form"`
		Files    map[string]string   `json:"files"`
		RawFiles map[string]string   `json:"raw_files"`
	}
)

var prefixMap = map[string]func(*http.Request, *request) error{
	"application/x-www-form-urlencoded": populateForm,
	"multipart/form-data":               populateMultipart,
}

func newServer(maxRequestLog int) *server {
	return &server{
		maxRequestLog: maxRequestLog,
		handlers:      []handler{},
		requests:      []*request{},
		logger:        log.New(os.Stdout, "[derision] ", 0),
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

	s.addHandler(handler)
	return response.Empty(http.StatusNoContent)
}

func (s *server) addHandler(handler handler) {
	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	defer s.mutex.Unlock()
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

	// Get a local ref to the request log so if we
	// clear it in the lines below we still have
	// something to serialize.
	requests := s.requests

	// Drop the request log
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

	s.addRequest(req)

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

func (s *server) addRequest(req *request) {
	s.mutex.Lock()
	s.requests = append(s.requests, req)
	s.pruneRequests()
	s.mutex.Unlock()
}

func (s *server) pruneRequests() {
	if s.maxRequestLog == 0 {
		return
	}

	for len(s.requests) > s.maxRequestLog {
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
// Snapshot

func convertRequest(r *http.Request) (*request, error) {
	snapshot := &request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header,
	}

	for prefix, fn := range prefixMap {
		if strings.HasPrefix(r.Header.Get("Content-Type"), prefix) {
			return snapshot, fn(r, snapshot)
		}
	}

	return snapshot, populateBody(r, snapshot)
}

func populateForm(r *http.Request, snapshot *request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	snapshot.Form = r.PostForm
	snapshot.Files = map[string]string{}
	snapshot.RawFiles = map[string]string{}
	snapshot.Body = ""
	snapshot.RawBody = ""
	return nil
}

func populateMultipart(r *http.Request, snapshot *request) error {
	if err := r.ParseMultipartForm(2e7); err != nil {
		return err
	}

	var (
		files    = map[string]string{}
		rawFiles = map[string]string{}
	)

	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				return err
			}

			defer file.Close()

			content, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}

			files[header.Filename] = string(content)
			rawFiles[header.Filename] = base64.StdEncoding.EncodeToString(content)
		}
	}

	snapshot.Files = files
	snapshot.RawFiles = rawFiles
	snapshot.Form = r.MultipartForm.Value
	snapshot.Body = ""
	snapshot.RawBody = ""
	return nil
}

func populateBody(r *http.Request, snapshot *request) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	snapshot.Files = map[string]string{}
	snapshot.RawFiles = map[string]string{}
	snapshot.Form = map[string][]string{}
	snapshot.Body = string(data)
	snapshot.RawBody = base64.StdEncoding.EncodeToString(data)
	return nil
}

//
// Helpers

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
