package template

import (
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/expectation"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
	. "github.com/onsi/gomega"
)

type SerializationSuite struct{}

func (s *SerializationSuite) TestUnmarshal(t sweet.T) {
	tmpl, err := Unmarshal([]byte(`{
		"status_code": "{{index .PathGroups 1}}",
		"headers": {
			"X-A": ["{{index .Headers \"X-A\" 0}}"],
			"X-B": ["{{index .Headers \"X-B\" 0}}"],
			"X-C": ["{{index .Headers \"X-C\" 0}}"]
		},
		"body": "{{.Method}} {{.Path}} :: {{.Body}}"
	}`))

	Expect(err).To(BeNil())

	r := &request.Request{
		Method: "GET",
		Path:   "/status/202",
		Headers: map[string][]string{
			"X-A": []string{"foo"},
			"X-B": []string{"bar"},
			"X-C": []string{"baz"},
		},
		Body: "foobar",
	}

	resp, err := tmpl.Respond(r, &expectation.Match{
		PathGroups: []string{"/status/202", "202"},
	})

	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusAccepted))

	headers, body, err := response.Serialize(resp)
	Expect(err).To(BeNil())
	Expect(headers).To(Equal(http.Header{
		"X-A":            []string{"foo"},
		"X-B":            []string{"bar"},
		"X-C":            []string{"baz"},
		"Content-Length": []string{"25"},
	}))
	Expect(body).To(Equal([]byte("GET /status/202 :: foobar")))

}

func (s *SerializationSuite) TestBadJSON(t sweet.T) {
	_, err := Unmarshal([]byte(``))
	Expect(err).To(MatchError("failed to unmarshal payload (unexpected end of JSON input)"))
}

func (s *SerializationSuite) TestBadStatusCodeTemplate(t sweet.T) {
	_, err := Unmarshal([]byte(`{"status_code": "{{"}`))
	Expect(err).To(MatchError("illegal status code template"))
}

func (s *SerializationSuite) TestHeaderPathTemplate(t sweet.T) {
	_, err := Unmarshal([]byte(`{"headers": {"Authorization": ["{{"]}}`))
	Expect(err).To(MatchError("illegal header template"))
}

func (s *SerializationSuite) TestBadBodyTemplate(t sweet.T) {
	_, err := Unmarshal([]byte(`{"body": "{{"}`))
	Expect(err).To(MatchError("illegal body template"))
}
