package server

import (
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/handler"
	"github.com/efritz/derision/internal/request"
	. "github.com/onsi/gomega"
)

type SerializationSuite struct{}

func init() {
	schemaPath = "../../schemas"
}

func (s *SerializationSuite) TestMakeHandler(t sweet.T) {
	handler, err := makeHandler([]byte(`{
		"request": {
			"method": "POST",
			"path": "/test"
		},
		"response": {
			"body": "ok"
		}
	}`))

	Expect(err).To(BeNil())

	// Matching request
	resp, err := handler(&request.Request{Method: "POST", Path: "/test"})
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusOK))

	// Non-matching request
	resp, err = handler(&request.Request{Method: "GET", Path: "/test"})
	Expect(err).To(BeNil())
	Expect(resp).To(BeNil())
}

func (s *SerializationSuite) TestMakeHandlerBadRequest(t sweet.T) {
	_, err := makeHandler([]byte(`{
		"request": {
			"method": "("
		},
		"response": {}
	}`))

	Expect(err).To(MatchError("failed to unmarshal expectation (illegal method regex)"))
}

func (s *SerializationSuite) TestMakeHandlerBadResponse(t sweet.T) {
	_, err := makeHandler([]byte(`{
		"request": {},
		"response": {
			"status_code": "{{"
		}
	}`))

	Expect(err).To(MatchError("failed to unmarshal template (illegal status code template)"))
}

func (s *SerializationSuite) TestMakeHandlerError(t sweet.T) {
	handler, err := makeHandler([]byte(`{
		"request": {},
		"response": {
			"body": "{{index .BodyGroups 3}}"
		}
	}`))

	Expect(err).To(BeNil())
	_, err = handler(&request.Request{Method: "POST", Path: "/test"})
	Expect(err).NotTo(BeNil())
}

func (s *SerializationSuite) TestMakeHandlersFromPath(t sweet.T) {
	handlers := handler.NewHandlerSet()
	err := loadHandlers(handlers, "./tests/valid")
	Expect(err).To(BeNil())

	resp, err := handlers.Handle(&request.Request{Method: "GET", Path: "/a1"})
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))

	resp, err = handlers.Handle(&request.Request{Method: "GET", Path: "/b2"})
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusOK))

	resp, err = handlers.Handle(&request.Request{Method: "POST", Path: "/d1"})
	Expect(err).To(BeNil())
	Expect(resp).To(BeNil())
}

func (s *SerializationSuite) TestMakeHandlersFromPathInvalidSchema(t sweet.T) {
	handlers := handler.NewHandlerSet()
	err := loadHandlers(handlers, "./tests/invalid-schema")
	Expect(err).To(MatchError("failed to load handlers from bad.yaml (invalid config: 0: response is required)"))
}

func (s *SerializationSuite) TestMakeHandlersFromPathInvalidTemplate(t sweet.T) {
	handlers := handler.NewHandlerSet()
	err := loadHandlers(handlers, "./tests/invalid-template")
	Expect(err).To(MatchError("failed to load handlers from bad.yaml (failed to unmarshal template (illegal body template))"))
}

func (s *SerializationSuite) TestMakeHandlersFromPathInvalidYAML(t sweet.T) {
	handlers := handler.NewHandlerSet()
	err := loadHandlers(handlers, "./tests/invalid-yaml")
	Expect(err).To(MatchError("failed to load handlers from bad.yaml (yaml: line 1: did not find expected node content)"))
}

func (s *SerializationSuite) TestMakeHandlersFromPathInvalidPath(t sweet.T) {
	handlers := handler.NewHandlerSet()
	err := loadHandlers(handlers, "./tests/missing")
	Expect(err).To(MatchError("failed to read config directory"))
}
