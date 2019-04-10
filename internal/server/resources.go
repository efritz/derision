package server

import (
	"context"
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/derision/internal/handler"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/nacelle"
	"github.com/efritz/response"
	"github.com/efritz/sse"
)

type (
	BaseResource struct {
		*chevron.EmptySpec
		HandlerSet handler.HandlerSet `service:"handler-set"`
		RequestLog request.Log        `service:"request-log"`
	}

	CatchAllHandler  struct{ *BaseResource }
	RegisterResource struct{ *BaseResource }
	ClearResource    struct{ *BaseResource }
	RequestsResource struct{ *BaseResource }

	SSEResource struct {
		*BaseResource
		sseServer *sse.Server
	}
)

func (r *CatchAllHandler) Handle(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	reqModel, err := convertRequest(req)
	if err != nil {
		logger.Error(err.Error())
		return response.Empty(http.StatusInternalServerError)
	}

	r.RequestLog.Add(reqModel)

	resp, err := r.HandlerSet.Handle(reqModel)
	if err != nil {
		logger.Error(err.Error())
		return response.Empty(http.StatusInternalServerError)
	}

	if resp == nil {
		resp = response.Empty(http.StatusNotFound)
	}

	return resp
}

func (r *RegisterResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	handler, err := makeHandler(middleware.GetJSONData(ctx))
	if err != nil {
		logger.Error(err.Error())
		return response.Empty(http.StatusInternalServerError)
	}

	r.HandlerSet.Add(handler)
	return response.Empty(http.StatusNoContent)
}

func (r *ClearResource) Post(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	r.HandlerSet.Clear()
	return response.Empty(http.StatusNoContent)
}

func (r *RequestsResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	return response.JSON(r.RequestLog.Copy(req.URL.Query().Get("clear") != ""))
}

func (r *SSEResource) PostInject() error {
	ch := make(chan interface{})

	go func() {
		for request := range r.RequestLog.Chan() {
			ch <- request
		}
	}()

	r.sseServer = sse.NewServer(ch)
	go r.sseServer.Start()
	return nil
}

func (r *SSEResource) Get(ctx context.Context, req *http.Request, logger nacelle.Logger) response.Response {
	return r.sseServer.Handler(req)
}
