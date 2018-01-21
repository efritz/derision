package main

import (
	"regexp"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ExpectationSuite struct{}

var TestExpectation = &expectation{
	method: regexp.MustCompile(`^GET$`),
	path:   regexp.MustCompile(`^/xyz/(\d+)$`),
	headers: map[string]*regexp.Regexp{
		"X-Secret":     regexp.MustCompile(`^sssh$`),
		"Content-Type": regexp.MustCompile(`^application/(json|xml)$`),
	},
	body: regexp.MustCompile(`"xyz_id": (\d+)`),
}

func (s *ExpectationSuite) TestSuccessfulMatch(t sweet.T) {
	match := TestExpectation.Matches(&request{
		Method: "GET",
		Path:   "/xyz/123",
		Headers: map[string][]string{
			"Content-Type": []string{"application/json"},
			"X-Secret":     []string{"sssh", "ignored value"},
			"X-Extra":      []string{"foo", "bar", "baz"},
		},
		Body: `{"xyz_id": 42, "zyx_id": 24}`,
	})

	Expect(match).NotTo(BeNil())
	Expect(match.methodGroups).To(Equal([]string{"GET"}))
	Expect(match.pathGroups).To(Equal([]string{"/xyz/123", "123"}))
	Expect(match.headerGroups).To(Equal(map[string][]string{
		"Content-Type": []string{"application/json", "json"},
		"X-Secret":     []string{"sssh"},
	}))
	Expect(match.bodyGroups).To(Equal([]string{`"xyz_id": 42`, "42"}))
}

func (s *ExpectationSuite) TestMatchFailsOnMethod(t sweet.T) {
	Expect(TestExpectation.Matches(&request{
		Method: "POST",
		Path:   "/xyz/123",
		Headers: map[string][]string{
			"Content-Type": []string{"application/json"},
			"X-Secret":     []string{"sssh"},
		},
		Body: `{"xyz_id": 42, "zyx_id": 24}`,
	})).To(BeNil())
}

func (s *ExpectationSuite) TestMatchFailsOnPath(t sweet.T) {
	Expect(TestExpectation.Matches(&request{
		Method: "GET",
		Path:   "/xyz/abc",
		Headers: map[string][]string{
			"Content-Type": []string{"application/json"},
			"X-Secret":     []string{"sssh"},
		},
		Body: `{"xyz_id": 42, "zyx_id": 24}`,
	})).To(BeNil())
}

func (s *ExpectationSuite) TestMatchFailsOnHeader(t sweet.T) {
	Expect(TestExpectation.Matches(&request{
		Method: "GET",
		Path:   "/xyz/123",
		Headers: map[string][]string{
			"Content-Type": []string{"application/yaml"},
			"X-Secret":     []string{"sssh"},
		},
		Body: `{"xyz_id": 42, "zyx_id": 24}`,
	})).To(BeNil())
}

func (s *ExpectationSuite) TestMatchFailsOnBody(t sweet.T) {
	Expect(TestExpectation.Matches(&request{
		Method: "GET",
		Path:   "/xyz/123",
		Headers: map[string][]string{
			"Content-Type": []string{"application/json"},
			"X-Secret":     []string{"sssh"},
		},
		Body: `{"xyz_id": "abc", "zyx_id": 24}`,
	})).To(BeNil())
}

func (s *ExpectationSuite) TestMatchRegexNilPattern(t sweet.T) {
	for _, val := range []string{"", "foo", "123"} {
		matched, matches := matchRegex(nil, val)
		Expect(matched).To(BeTrue())
		Expect(matches).To(BeEmpty())
	}
}

func (s *ExpectationSuite) TestMatchRegexMatch(t sweet.T) {
	re := regexp.MustCompile(`^abc\d+xyz$`)

	matched, matches := matchRegex(re, "abc123xyz")
	Expect(matched).To(BeTrue())
	Expect(matches).To(Equal([]string{"abc123xyz"}))
}

func (s *ExpectationSuite) TestMatchRegexNoMatch(t sweet.T) {
	re := regexp.MustCompile(`^abc\d+xyz$`)

	matched, matches := matchRegex(re, "abc123")
	Expect(matched).To(BeFalse())
	Expect(matches).To(BeEmpty())
}

func (s *ExpectationSuite) TestMatchRegexMatchWithGroups(t sweet.T) {
	re := regexp.MustCompile(`^abc(\d+)xyz$`)

	matched, matches := matchRegex(re, "abc123xyz")
	Expect(matched).To(BeTrue())
	Expect(matches).To(Equal([]string{"abc123xyz", "123"}))
}

func (s *ExpectationSuite) TestGetFirst(t sweet.T) {
	Expect(getFirst(nil, "foo")).To(BeEmpty())
	Expect(getFirst(map[string][]string{"foo": []string{"bar"}}, "foo")).To(Equal("bar"))
	Expect(getFirst(map[string][]string{"foo": []string{"bar", "baz"}}, "foo")).To(Equal("bar"))
}
