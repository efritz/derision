package template

import (
	"net/http"
	tmpl "text/template"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/expectation"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
	. "github.com/onsi/gomega"
)

type TemplateSuite struct{}

func (s *TemplateSuite) TestRespond(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(`{{index .PathGroups 1}}`),
		headers: map[string][]*tmpl.Template{
			"X-A": []*tmpl.Template{testCompile(`{{index .Headers "X-A" 0}}`)},
			"X-B": []*tmpl.Template{testCompile(`{{index .Headers "X-B" 0}}`)},
			"X-C": []*tmpl.Template{testCompile(`{{index .Headers "X-C" 0}}`)},
		},
		body: testCompile(`{{.Method}} {{.Path}} :: {{.Body}}`),
	}

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

func (s *TemplateSuite) TestRespondEmptyStatusCode(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(``),
		body:       testCompile(`test`)}

	resp, err := tmpl.Respond(&request.Request{}, &expectation.Match{})
	Expect(err).To(BeNil())
	Expect(resp.StatusCode()).To(Equal(http.StatusOK))
}

func (s *TemplateSuite) TestRespondIllegalStatusCode(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(`abc`),
		body:       testCompile(`test`)}

	_, err := tmpl.Respond(&request.Request{}, &expectation.Match{})
	Expect(err).NotTo(BeNil())
	Expect(err).To(Equal(ErrIllegalStatusCode))
}

func testCompile(text string) *tmpl.Template {
	return tmpl.Must(tmpl.New("test").Parse(text))
}

func (s *TemplateSuite) TestRespondBadBodyTemplateApplication(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(``),
		headers:    map[string][]*tmpl.Template{},
		body:       testCompile(`{{index .Headers "missing" 0}}`),
	}

	_, err := tmpl.Respond(&request.Request{}, &expectation.Match{})
	Expect(err).NotTo(BeNil())
}

func (s *TemplateSuite) TestRespondStatusCodeTemplateApplication(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(`{{index .Headers "missing" 0}}`),
		headers:    map[string][]*tmpl.Template{},
		body:       testCompile(``),
	}

	_, err := tmpl.Respond(&request.Request{}, &expectation.Match{})
	Expect(err).NotTo(BeNil())
}

func (s *TemplateSuite) TestRespondHeaderTemplateApplication(t sweet.T) {
	tmpl := &template{
		statusCode: testCompile(``),
		headers: map[string][]*tmpl.Template{
			"X-Test": []*tmpl.Template{testCompile(`{{index .Headers "missing" 0}}`)},
		},
		body: testCompile(``),
	}

	_, err := tmpl.Respond(&request.Request{}, &expectation.Match{})
	Expect(err).NotTo(BeNil())
}
