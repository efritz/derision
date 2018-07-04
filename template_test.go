package main

import (
	"net/http"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type TemplateSuite struct{}

func (s *TemplateSuite) TestRespond(t sweet.T) {
	template := &template{
		StatusCode: `200`,
		Headers: map[string][]string{
			"X-Test": []string{"yes"},
			"X-Path": []string{`{{ .Path }}`},
			"X-Meth": []string{`{{ .Method }}`},
			"X-Type": []string{`{{ index (index .HeaderGroups "Content-Type") 1 }}`},
		},
		Body: `You requested details for user {{ index .PathGroups 1 }}.`,
	}

	r := &request{
		Method:  "GET",
		Path:    "/users/123",
		Headers: map[string][]string{"Content-Type": []string{"application/json"}},
		Body:    "",
	}

	m := &match{
		methodGroups: []string{"GET"},
		pathGroups:   []string{"/users/123", "123"},
		headerGroups: map[string][]string{"Content-Type": []string{"application/json", "json"}},
	}

	resp, err := template.Respond(r, m)
	Expect(err).To(BeNil())

	Expect(resp.StatusCode()).To(Equal(http.StatusOK))
	Expect(resp.Header("X-Test")).To(Equal("yes"))
	Expect(resp.Header("X-Path")).To(Equal("/users/123"))
	Expect(resp.Header("X-Meth")).To(Equal("GET"))
	Expect(resp.Header("X-Type")).To(Equal("json"))
	Expect(getBody(resp)).To(Equal("You requested details for user 123."))
}

func (s *TemplateSuite) TestApplyTemplate(t sweet.T) {
	val, err := applyTemplate("", `a {{ .A }} {{ index .B "bar" }} z`, map[string]interface{}{
		"A": "foo",
		"B": map[string]int{"bar": 27},
	})

	Expect(err).To(BeNil())
	Expect(val).To(Equal("a foo 27 z"))
}

func (s *TemplateSuite) TestApplyTemplateError(t sweet.T) {
	_, err := applyTemplate("", `a {{ .B }} {{ index .A "bar" }} z`, map[string]interface{}{
		"A": "foo",
		"B": map[string]int{"bar": 27},
	})

	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(ContainSubstring("bar"))
}
