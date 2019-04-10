package handler

import (
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/response"
)

type Handler func(r *request.Request) (response.Response, error)
