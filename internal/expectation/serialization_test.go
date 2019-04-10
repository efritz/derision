package expectation

import (
	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/request"
	. "github.com/onsi/gomega"
)

type SerializationSuite struct{}

func (s *SerializationSuite) TestUnmarshal(t sweet.T) {
	e, err := Unmarshal([]byte(`{
		"method": "GET|POST",
		"path": "/(.*)",
		"headers": {
			"Authorization": "Basic (.*)"
		},
		"body": ".*"
	}`))

	Expect(err).To(BeNil())

	match := e.Matches(&request.Request{
		Method: "GET",
		Path:   "/users/123",
		Headers: map[string][]string{
			"Authorization": []string{"Basic secret"},
		},
		Body: "foobar",
	})

	Expect(match).NotTo(BeNil())
	Expect(match.MethodGroups).To(Equal([]string{"GET"}))
	Expect(match.PathGroups).To(Equal([]string{"/users/123", "users/123"}))
	Expect(match.HeaderGroups).To(Equal(map[string][]string{
		"Authorization": []string{"Basic secret", "secret"},
	}))
	Expect(match.BodyGroups).To(Equal([]string{"foobar"}))
}

func (s *SerializationSuite) TestUnmarshalEmpty(t sweet.T) {
	e1, err := Unmarshal([]byte(`{"method": "GET|POST"}`))
	Expect(err).To(BeNil())

	e2, err := Unmarshal([]byte(`{"path": "/(.*)"}`))
	Expect(err).To(BeNil())

	e3, err := Unmarshal([]byte(`{"headers": {"Authorization": "Basic (.*)"}}`))
	Expect(err).To(BeNil())

	e4, err := Unmarshal([]byte(`{"body": "foobar"}`))
	Expect(err).To(BeNil())

	for _, e := range []Expectation{e1, e2, e3, e4} {
		match := e.Matches(&request.Request{
			Method: "GET",
			Path:   "/users/123",
			Headers: map[string][]string{
				"Authorization": []string{"Basic secret"},
			},
			Body: "foobar",
		})

		Expect(match).NotTo(BeNil())
	}
}

func (s *SerializationSuite) TestBadJSON(t sweet.T) {
	_, err := Unmarshal([]byte(``))
	Expect(err).To(MatchError("failed to unmarshal payload (unexpected end of JSON input)"))
}

func (s *SerializationSuite) TestBadMethodRegex(t sweet.T) {
	_, err := Unmarshal([]byte(`{"method": "("}`))
	Expect(err).To(MatchError("illegal method regex"))
}

func (s *SerializationSuite) TestBadPathRegex(t sweet.T) {
	_, err := Unmarshal([]byte(`{"path": "("}`))
	Expect(err).To(MatchError("illegal path regex"))
}

func (s *SerializationSuite) TestHeaderPathRegex(t sweet.T) {
	_, err := Unmarshal([]byte(`{"headers": {"Authorization": "("}}`))
	Expect(err).To(MatchError("illegal header regex"))
}

func (s *SerializationSuite) TestBadBodyRegex(t sweet.T) {
	_, err := Unmarshal([]byte(`{"body": "("}`))
	Expect(err).To(MatchError("illegal body regex"))
}
