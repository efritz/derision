package handler

import (
	"fmt"
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
	. "github.com/onsi/gomega"
)

type SetSuite struct{}

func (s *SetSuite) TestHandle(t sweet.T) {
	set := NewHandlerSet()
	set.Add(makeHandler("/foo", http.StatusOK))
	set.Add(makeHandler("/bar", http.StatusNotFound))
	set.Add(makeHandler("/baz", http.StatusConflict))

	resp, err := set.Handle(&request.Request{Path: "/foo"})
	Expect(err).To(BeNil())
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode()).To(Equal(200))

	resp, err = set.Handle(&request.Request{Path: "/bar"})
	Expect(err).To(BeNil())
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode()).To(Equal(404))

	resp, err = set.Handle(&request.Request{Path: "/baz"})
	Expect(err).To(BeNil())
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode()).To(Equal(409))

	resp, err = set.Handle(&request.Request{Path: "/bonk"})
	Expect(err).To(BeNil())
	Expect(resp).To(BeNil())
}

func (s *SetSuite) TestHandleError(t sweet.T) {
	set := NewHandlerSet()

	set.Add(func(r *request.Request) (response.Response, error) {
		return nil, fmt.Errorf("oops")
	})

	_, err := set.Handle(&request.Request{Path: "/foo"})
	Expect(err).To(MatchError("oops"))
}

func (s *SetSuite) TestHandleClear(t sweet.T) {
	set := NewHandlerSet()
	set.Add(makeHandler("/foo", http.StatusOK))

	resp, err := set.Handle(&request.Request{Path: "/foo"})
	Expect(err).To(BeNil())
	Expect(resp).NotTo(BeNil())
	Expect(resp.StatusCode()).To(Equal(200))

	set.Clear()
	resp, err = set.Handle(&request.Request{Path: "/foo"})
	Expect(err).To(BeNil())
	Expect(resp).To(BeNil())
}

func makeHandler(path string, status int) Handler {
	return func(r *request.Request) (response.Response, error) {
		if r.Path == path {
			return response.Empty(status), nil
		}

		return nil, nil
	}
}
