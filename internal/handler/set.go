package handler

import (
	"sync"

	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
)

type (
	HandlerSet interface {
		Handle(r *request.Request) (response.Response, error)
		Add(handler Handler)
		Clear()
	}

	handlerSet struct {
		handlers []Handler
		mutex    sync.RWMutex
	}
)

func NewHandlerSet() *handlerSet {
	return &handlerSet{}
}

func (s *handlerSet) Handle(r *request.Request) (response.Response, error) {
	for _, handler := range s.handlers {
		if resp, err := handler(r); err != nil || resp != nil {
			return resp, err
		}
	}

	return nil, nil
}

func (s *handlerSet) Add(handler Handler) {
	s.mutex.Lock()
	s.handlers = append(s.handlers, handler)
	s.mutex.Unlock()
}

func (s *handlerSet) Clear() {
	s.mutex.Lock()
	s.handlers = s.handlers[:0]
	s.mutex.Unlock()
}
