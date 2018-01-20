package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/efritz/response"
	. "github.com/onsi/gomega"
)

type ServerSuite struct{}

func (s *ServerSuite) TestRegisterHandler(t sweet.T) {
	server := newServer()
	resp := server.registerHandler(makeRequest("POST", "http://deriosio.io/_control/register", bytes.NewBuffer([]byte(`{
		"request": {"method": "POST"},
		"response": {"body": "foo"}
	}`))))

	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
	Expect(server.handlers).To(HaveLen(1))
	testHandlerBehavior(server.handlers[0])
}

func (s *ServerSuite) TestClearHandler(t sweet.T) {
	server := newServer()
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, NoopHandler)

	resp := server.clearHandler(makeRequest("GET", "http://derision.io/_control/clear", nil))
	Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
	Expect(server.handlers).To(BeEmpty())
}

func (s *ServerSuite) TestRequestsHandler(t sweet.T) {
	server := newServer()
	server.apiHandler(makeRequest("GET", "http://derision.io/a", bytes.NewReader([]byte("foo"))))
	server.apiHandler(makeRequest("GET", "http://derision.io/b", bytes.NewReader([]byte("bar"))))
	server.apiHandler(makeRequest("GET", "http://derision.io/c", bytes.NewReader([]byte("baz"))))

	body := getRequestsBody(server, "http://derision.io/_control/requests")

	Expect(body).To(HaveLen(3))
	Expect(body[0]["path"]).To(Equal("/a"))
	Expect(body[1]["path"]).To(Equal("/b"))
	Expect(body[2]["path"]).To(Equal("/c"))
	Expect(body[0]["body"]).To(Equal("foo"))
	Expect(body[1]["body"]).To(Equal("bar"))
	Expect(body[2]["body"]).To(Equal("baz"))
}

func (s *ServerSuite) TestRequestsHandlerClear(t sweet.T) {
	server := newServer()
	server.apiHandler(makeRequest("GET", "http://derision.io/a", bytes.NewReader([]byte("foo"))))
	server.apiHandler(makeRequest("GET", "http://derision.io/b", bytes.NewReader([]byte("bar"))))
	server.apiHandler(makeRequest("GET", "http://derision.io/c", bytes.NewReader([]byte("baz"))))

	Expect(getRequestsBody(server, "http://derision.io/_control/requests")).To(HaveLen(3))
	Expect(getRequestsBody(server, "http://derision.io/_control/requests?clear=true")).To(HaveLen(3))
	Expect(getRequestsBody(server, "http://derision.io/_control/requests")).To(HaveLen(0))
}

func (s *ServerSuite) TestAPIHandler(t sweet.T) {
	server := newServer()
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, ConditionalHandler)
	server.handlers = append(server.handlers, ErrorHandler)

	resp := server.apiHandler(makeRequest("GET", "/xyz", bytes.NewReader(nil)))
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(getBody(resp)).To(Equal(`["foo","bar","baz"]`))
}

func (s *ServerSuite) TestAPIHandlerError(t sweet.T) {
	server := newServer()
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, ConditionalHandler)
	server.handlers = append(server.handlers, ErrorHandler)

	resp := server.apiHandler(makeRequest("GET", "/wxy", bytes.NewReader(nil)))
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
}

func (s *ServerSuite) TestAPIHandlerNotFound(t sweet.T) {
	server := newServer()
	server.handlers = append(server.handlers, NoopHandler)
	server.handlers = append(server.handlers, ConditionalHandler)
	server.handlers = append(server.handlers, NoopHandler)

	resp := server.apiHandler(makeRequest("GET", "/wxy", bytes.NewReader(nil)))
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
}

func (s *ServerSuite) TestConvertRequest(t sweet.T) {
	r := makeRequest("POST", "http://derision.io:8080/xyz/123", bytes.NewReader([]byte("foo\nbar\nbaz\n")))
	r.Header.Set("X-Foo", "bar")
	r.Header.Set("X-Bar", "baz")

	req, err := convertRequest(r)
	Expect(err).To(BeNil())

	Expect(req.Method).To(Equal("POST"))
	Expect(req.Path).To(Equal("/xyz/123"))
	Expect(req.Body).To(Equal("foo\nbar\nbaz\n"))
	Expect(req.Headers).To(Equal(map[string][]string{
		"X-Foo": []string{"bar"},
		"X-Bar": []string{"baz"},
	}))
}

func (s *ServerSuite) TestConvertRequestReadError(t sweet.T) {
	_, err := convertRequest(makeRequest("POST", "http://derision.io:8080/xyz/123", &failingReader{}))
	Expect(err).NotTo(BeNil())
}

func (s *ServerSuite) TestMakeHandler(t sweet.T) {
	handler, err := makeHandler(ioutil.NopCloser(bytes.NewBuffer([]byte(`{
		"request": {"method": "POST"},
		"response": {"body": "foo"}
	}`))))

	Expect(err).To(BeNil())
	testHandlerBehavior(handler)
}

func (s *ServerSuite) TestMakeHandlerError(t sweet.T) {
	_, err := makeHandler(ioutil.NopCloser(bytes.NewBuffer([]byte("null"))))
	Expect(err).NotTo(BeNil())
}

//
// Common Testing

func testHandlerBehavior(handler handler) {
	// No match
	resp, err := handler(&request{Method: "GET", Path: "/xyz", Body: ""})
	Expect(err).To(BeNil())
	Expect(resp).To(BeNil())

	// Match
	resp, err = handler(&request{Method: "POST", Path: "/xyz", Body: ""})
	Expect(err).To(BeNil())
	Expect(resp).NotTo(BeNil())
	Expect(getBody(resp)).To(Equal("foo"))
}

//
// Helpers

func getBody(resp *response.Response) string {
	data, _ := response.Serialize(resp)
	return string(data)
}

func getRequestsBody(server *server, url string) []map[string]interface{} {
	resp := server.requestsHandler(makeRequest("GET", url, nil))

	body := []map[string]interface{}{}
	json.Unmarshal([]byte(getBody(resp)), &body)
	return body
}

func makeRequest(method, url string, body io.Reader) *http.Request {
	r, _ := http.NewRequest(method, url, body)
	return r
}

//
// Mocks

func NoopHandler(r *request) (*response.Response, error) {
	return nil, nil
}

func ErrorHandler(r *request) (*response.Response, error) {
	return nil, fmt.Errorf("utoh")
}

func ConditionalHandler(r *request) (*response.Response, error) {
	if r.Path == "/xyz" {
		return response.JSON([]string{"foo", "bar", "baz"}), nil
	}

	return nil, nil
}

type failingReader struct{}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("utoh")
}
