package server

import (
	"context"
	"net/http"

	"github.com/aphistic/sweet"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	. "github.com/onsi/gomega"
)

type MiddlewareSuite struct{}

func (s *MiddlewareSuite) TestConvert(t sweet.T) {
	catchAllHandler := func(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
		return response.Empty(http.StatusCreated)
	}

	wrappedHandler := func(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
		return response.Empty(http.StatusAccepted)
	}

	handler, err := NewControlMiddleware(catchAllHandler).Convert(wrappedHandler)
	Expect(err).To(BeNil())

	r1, _ := http.NewRequest("GET", "http://test.io", nil)
	r2, _ := http.NewRequest("GET", "http://test.io", nil)
	r2.Header.Set("X-Derision-Control", "true")

	resp1 := handler(context.Background(), r1, nacelle.NewNilLogger())
	resp2 := handler(context.Background(), r2, nacelle.NewNilLogger())

	Expect(resp1.StatusCode()).To(Equal(http.StatusCreated))
	Expect(resp2.StatusCode()).To(Equal(http.StatusAccepted))
}
