package server

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/request"
	. "github.com/onsi/gomega"
)

type ConversionSuite struct{}

var multipartBody = `
--xxx
Content-Disposition: form-data; name="field1"

value1
--xxx
Content-Disposition: form-data; name="field2"

value2
--xxx
Content-Disposition: form-data; name="file"; filename="file"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

binary data
--xxx--
`

func (s *ConversionSuite) TestConvertSimple(t sweet.T) {
	body := bytes.NewReader([]byte("foo\nbar\nbaz\n"))
	r, err := http.NewRequest("POST", "http://test.io/path", body)
	Expect(err).To(BeNil())
	r.Header.Add("X-Foo", "bar")
	r.Header.Add("X-Bar", "baz")
	r.Header.Add("X-Bar", "bonk")

	converted, err := convertRequest(r)
	Expect(err).To(BeNil())
	Expect(converted).To(Equal(&request.Request{
		Method: "POST",
		Path:   "/path",
		Headers: map[string][]string{
			"X-Foo": []string{"bar"},
			"X-Bar": []string{"baz", "bonk"},
		},
		Body:     "foo\nbar\nbaz\n",
		RawBody:  "Zm9vCmJhcgpiYXoK",
		Form:     map[string][]string{},
		Files:    map[string]string{},
		RawFiles: map[string]string{},
	}))
}

func (s *ConversionSuite) TestConvertForm(t sweet.T) {
	body := bytes.NewReader([]byte("z=post&both=y"))
	r, err := http.NewRequest("POST", "http://test.io/path?q=foo&q=bar&both=x", body)
	Expect(err).To(BeNil())
	r.Header.Add("X-Foo", "bar")
	r.Header.Add("X-Bar", "baz")
	r.Header.Add("X-Bar", "bonk")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	converted, err := convertRequest(r)
	Expect(err).To(BeNil())
	Expect(converted).To(Equal(&request.Request{
		Method: "POST",
		Path:   "/path",
		Headers: map[string][]string{
			"X-Foo":        []string{"bar"},
			"X-Bar":        []string{"baz", "bonk"},
			"Content-Type": []string{"application/x-www-form-urlencoded; param=value"},
		},
		Body:    "z=post&both=y",
		RawBody: "ej1wb3N0JmJvdGg9eQ==",
		Form: map[string][]string{
			"q":    []string{"foo", "bar"},
			"z":    []string{"post"},
			"both": []string{"y", "x"},
		},
		Files:    map[string]string{},
		RawFiles: map[string]string{},
	}))
}

func (s *ConversionSuite) TestConvertMultipartForm(t sweet.T) {
	body := bytes.NewReader([]byte(multipartBody))
	r, err := http.NewRequest("POST", "http://test.io/path", body)
	Expect(err).To(BeNil())
	r.Header.Add("X-Foo", "bar")
	r.Header.Add("X-Bar", "baz")
	r.Header.Add("X-Bar", "bonk")
	r.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")

	r.Form = url.Values{}
	r.Form.Add("x", "123")
	r.Form.Add("y", "456")
	r.Form.Add("z", "789")

	converted, err := convertRequest(r)
	Expect(err).To(BeNil())
	Expect(converted).To(Equal(&request.Request{
		Method: "POST",
		Path:   "/path",
		Headers: map[string][]string{
			"X-Foo":        []string{"bar"},
			"X-Bar":        []string{"baz", "bonk"},
			"Content-Type": []string{"multipart/form-data; boundary=xxx"},
		},
		Body:    multipartBody,
		RawBody: "Ci0teHh4CkNvbnRlbnQtRGlzcG9zaXRpb246IGZvcm0tZGF0YTsgbmFtZT0iZmllbGQxIgoKdmFsdWUxCi0teHh4CkNvbnRlbnQtRGlzcG9zaXRpb246IGZvcm0tZGF0YTsgbmFtZT0iZmllbGQyIgoKdmFsdWUyCi0teHh4CkNvbnRlbnQtRGlzcG9zaXRpb246IGZvcm0tZGF0YTsgbmFtZT0iZmlsZSI7IGZpbGVuYW1lPSJmaWxlIgpDb250ZW50LVR5cGU6IGFwcGxpY2F0aW9uL29jdGV0LXN0cmVhbQpDb250ZW50LVRyYW5zZmVyLUVuY29kaW5nOiBiaW5hcnkKCmJpbmFyeSBkYXRhCi0teHh4LS0K",
		Form: map[string][]string{
			"x":      []string{"123"},
			"y":      []string{"456"},
			"z":      []string{"789"},
			"field1": []string{"value1"},
			"field2": []string{"value2"},
		},
		Files: map[string]string{
			"file": "binary data",
		},
		RawFiles: map[string]string{
			"file": "YmluYXJ5IGRhdGE=",
		},
	}))
}
