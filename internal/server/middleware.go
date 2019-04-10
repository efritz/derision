package server

import (
	"context"
	"net/http"

	"github.com/efritz/nacelle"
	"github.com/efritz/response"

	"github.com/efritz/chevron"
)

type ControlMiddleware struct {
	catchAllHandler chevron.Handler
}

func NewControlMiddleware(catchAllHandler chevron.Handler) chevron.Middleware {
	return &ControlMiddleware{
		catchAllHandler: catchAllHandler,
	}
}

func (m *ControlMiddleware) Convert(f chevron.Handler) (chevron.Handler, error) {
	handler := func(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
		if req.Header.Get("X-Derision-Control") != "" {
			return f(ctx, req, logger)
		}

		return m.catchAllHandler(ctx, req, logger)
	}

	return handler, nil
}
