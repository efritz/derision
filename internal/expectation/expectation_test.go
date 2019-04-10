package expectation

import (
	"regexp"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/request"
	. "github.com/onsi/gomega"
)

type ExpectationSuite struct{}

func (s *ExpectationSuite) TestMatchMethod(t sweet.T) {
	var match *Match
	e1 := &expectation{method: regexp.MustCompile("GET|POST")}
	e2 := &expectation{method: regexp.MustCompile("P(.*)")}

	// Without groups
	match = e1.Matches(&request.Request{Method: "GET"})
	Expect(match).NotTo(BeNil())
	Expect(match.MethodGroups).To(Equal([]string{"GET"}))

	// With groups
	match = e2.Matches(&request.Request{Method: "POST"})
	Expect(match).NotTo(BeNil())
	Expect(match.MethodGroups).To(Equal([]string{"POST", "OST"}))

	// No match
	match = e1.Matches(&request.Request{Method: "PATCH"})
	Expect(match).To(BeNil())
}

func (s *ExpectationSuite) TestMatchPath(t sweet.T) {
	var match *Match
	e1 := &expectation{path: regexp.MustCompile("/users")}
	e2 := &expectation{path: regexp.MustCompile("/users/(\\w+)")}

	// Without groups
	match = e1.Matches(&request.Request{Path: "/users"})
	Expect(match).NotTo(BeNil())
	Expect(match.PathGroups).To(Equal([]string{"/users"}))

	// With groups
	match = e2.Matches(&request.Request{Path: "/users/foobar"})
	Expect(match).NotTo(BeNil())
	Expect(match.PathGroups).To(Equal([]string{"/users/foobar", "foobar"}))

	// No match
	match = e1.Matches(&request.Request{Path: "/me"})
	Expect(match).To(BeNil())
}

func (s *ExpectationSuite) TestMatchHeader(t sweet.T) {
	r1 := regexp.MustCompile("\\d{4}-\\d{4}")
	r2 := regexp.MustCompile("\\d{4}-(\\d{4})")

	var match *Match
	e1 := &expectation{headers: map[string]*regexp.Regexp{"X-Test": r1}}
	e2 := &expectation{headers: map[string]*regexp.Regexp{"X-Test": r2}}

	// Without groups
	match = e1.Matches(&request.Request{Headers: map[string][]string{
		"X-Test": []string{"1234-5678"},
	}})

	Expect(match).NotTo(BeNil())
	Expect(match.HeaderGroups).To(Equal(map[string][]string{
		"X-Test": []string{"1234-5678"},
	}))

	// With groups
	match = e2.Matches(&request.Request{Headers: map[string][]string{
		"X-Test": []string{"1234-5678"},
	}})

	Expect(match).NotTo(BeNil())
	Expect(match.HeaderGroups).To(Equal(map[string][]string{
		"X-Test": []string{"1234-5678", "5678"},
	}))

	// No match (bad value)
	match = e1.Matches(&request.Request{Headers: map[string][]string{
		"X-Test": []string{"abcd-efgh"},
	}})

	Expect(match).To(BeNil())

	// No match (missing header)
	match = e1.Matches(&request.Request{Headers: map[string][]string{
		"Y-Test": []string{"PATCH"},
	}})

	Expect(match).To(BeNil())
}

func (s *ExpectationSuite) TestMatchBody(t sweet.T) {
	var match *Match
	e1 := &expectation{body: regexp.MustCompile("foo: bar")}
	e2 := &expectation{body: regexp.MustCompile("foo: (.*)")}

	// Without groups
	match = e1.Matches(&request.Request{Body: "foo: bar"})
	Expect(match).NotTo(BeNil())
	Expect(match.BodyGroups).To(Equal([]string{"foo: bar"}))

	// With groups
	match = e2.Matches(&request.Request{Body: "foo: bar"})
	Expect(match).NotTo(BeNil())
	Expect(match.BodyGroups).To(Equal([]string{"foo: bar", "bar"}))

	// No match
	match = e1.Matches(&request.Request{Body: "bar: foo"})
	Expect(match).To(BeNil())
}
