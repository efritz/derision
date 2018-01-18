package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/xeipuuv/gojsonschema"
)

type (
	server struct {
		mutex    sync.RWMutex
		handlers []handler
		requests []*request
	}

	request struct {
		Method  string              `json:"method"`
		URL     string              `json:"url"`
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	}

	requestExpectation struct {
		Method  string              `json:"method"`
		Path    string              `json:"path"`
		Headers map[string][]string `json:"headers"`
	}

	responseTemplate struct {
		StatusCode string              `json:"status_code"`
		Headers    map[string][]string `json:"headers"`
		Body       string              `json:"body"`
	}

	jsonHandler struct {
		RequestExpectation requestExpectation `json:"request"`
		ResponseTemplate   responseTemplate   `json:"response"`
	}

	handler func(w http.ResponseWriter, r *request) bool
)

const handlerSchema = `
{
	"type": "object",
	"properties": {
		"request": {
			"type": "object",
			"properties": {
				"method": {
					"type": "string",
					"enum": ["GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"]
				},
				"path": {
					"type": "string"
				},
				"headers": {
					"type": "object"
				}
			},
			"additionalProperties": false
		},
		"response": {
			"type": "object",
			"properties": {
				"status_code": {
					"type": "string",
					"pattern": "^\\d{3}$"
				},
				"headers": {
					"type": "object"
				},
				"body": {
					"type": "string"
				}
			},
			"additionalProperties": false
		}
	},
	"additionalProperties": false,
	"required": [
		"request",
		"response"
	]
}
`

func main() {
	negroni := negroni.Classic()
	negroni.UseHandler(makeRouter(&server{
		handlers: []handler{},
		requests: []*request{},
	}))

	negroni.Run("0.0.0.0:5000")
}

func makeRouter(s *server) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	r.PathPrefix("/_control").Path("/clear").Methods("POST").HandlerFunc(s.clearHandler)
	r.PathPrefix("/_control").Path("/register").Methods("POST").HandlerFunc(s.registerHandler)
	r.PathPrefix("/_control").Path("/gather").Methods("GET").HandlerFunc(s.gatherHandler)
	r.NotFoundHandler = http.HandlerFunc(s.apiHandler)

	return r
}

func (s *server) apiHandler(w http.ResponseWriter, r *http.Request) {
	req, err := convertRequest(r)
	if err != nil {
		fmt.Printf("Failed to convert request (%s)\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.mutex.Lock()
	s.requests = append(s.requests, req)
	s.mutex.Unlock()

	for _, handler := range s.handlers {
		if handler(w, req) {
			return
		}
	}

	fmt.Printf("Did not find a matching handler for request\n")
	w.WriteHeader(http.StatusNotFound)
}

func convertRequest(r *http.Request) (*request, error) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return &request{
		Method:  r.Method,
		URL:     r.URL.String(),
		Headers: r.Header,
		Body:    string(body),
	}, nil
}

func (s *server) clearHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	s.handlers = s.handlers[:0]
	s.mutex.Unlock()
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {
	handler, err := makeHandler(r.Body)
	if err != nil {
		fmt.Printf("Failed to make handler (%s)\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	s.mutex.Unlock()
}

func (e *requestExpectation) Matches(r *request) bool {
	if e.Method != "" && r.Method != e.Method {
		return false
	}

	if e.Path != "" && r.URL != e.Path {
		return false
	}

	for k, match := range e.Headers {
		vals, ok := r.Headers[k]
		if !ok {
			return false
		}

		if len(vals) != len(match) {
			return false
		}

		for i, v := range vals {
			if v != match[i] {
				return false
			}
		}
	}

	return true
}

func (t *responseTemplate) Write(w http.ResponseWriter, r *request) {
	for header, values := range t.Headers {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	if t.StatusCode == "" {
		w.WriteHeader(http.StatusOK)
	} else {
		code, err := strconv.Atoi(t.StatusCode)
		if err != nil {
			panic(err.Error())
		}

		w.WriteHeader(code)
	}

	chunk := []byte(t.Body)

	for len(chunk) > 0 {
		n, err := w.Write(chunk)
		if err != nil {
			panic(err.Error())
		}

		chunk = chunk[n:]
	}
}

func makeHandler(r io.ReadCloser) (handler, error) {
	reqExpectation, respTemplate, err := readAndValidate(r)
	if err != nil {
		return nil, err
	}

	handler := func(w http.ResponseWriter, r *request) bool {
		if reqExpectation.Matches(r) {
			respTemplate.Write(w, r)
			return true
		}

		return false
	}

	return handler, nil
}

func readAndValidate(r io.ReadCloser) (*requestExpectation, *responseTemplate, error) {
	defer r.Close()

	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewStringLoader(handlerSchema),
		gojsonschema.NewStringLoader(string(input)),
	)

	if err != nil {
		return nil, nil, err
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			fmt.Printf("Validation error (%s: %s)\n", err.Field(), err.Description())
		}

		return nil, nil, fmt.Errorf("validation error")
	}

	payload := &jsonHandler{}
	if err := json.Unmarshal(input, &payload); err != nil {
		return nil, nil, err
	}

	return &payload.RequestExpectation, &payload.ResponseTemplate, nil
}

func (s *server) gatherHandler(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := json.Marshal(s.requests)
	if err != nil {
		fmt.Printf("Failed to serialize request list (%s)\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Get("clear") != "" {
		s.requests = s.requests[:0]
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
