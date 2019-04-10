package server

import (
	"net/http"

	"github.com/efritz/chevron"
	"github.com/efritz/chevron/middleware"
	"github.com/efritz/derision/internal/handler"
	"github.com/efritz/derision/internal/request"
	"github.com/efritz/nacelle"
	basehttp "github.com/efritz/nacelle/base/http"
	"github.com/efritz/response"
	"github.com/xeipuuv/gojsonschema"
)

type Server struct {
	Services      nacelle.ServiceContainer `service:"container"`
	wrappedServer nacelle.Process
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Init(config nacelle.Config) error {
	if err := setupDataStructures(config, s.Services); err != nil {
		return err
	}

	catchAllHandler := &CatchAllHandler{}
	if err := s.Services.Inject(catchAllHandler); err != nil {
		return err
	}

	server, err := newServer(config, s.Services, catchAllHandler.Handle)
	if err != nil {
		return err
	}

	s.wrappedServer = server
	return nil
}

func (s *Server) Start() error {
	return s.wrappedServer.Start()
}

func (s *Server) Stop() error {
	return s.wrappedServer.Stop()
}

func setupDataStructures(config nacelle.Config, services nacelle.ServiceContainer) error {
	serverConfig := &Config{}
	if err := config.Load(serverConfig); err != nil {
		return err
	}

	handlerSet := handler.NewHandlerSet()
	requestLog := request.NewLog(serverConfig.RequestLogCapacity)

	if serverConfig.ConfigDir != "" {
		if err := loadHandlers(handlerSet, serverConfig.ConfigDir); err != nil {
			return err
		}
	}

	if err := services.Set("request-log", requestLog); err != nil {
		return err
	}

	if err := services.Set("handler-set", handlerSet); err != nil {
		return err
	}

	return nil
}

func newServer(
	config nacelle.Config,
	services nacelle.ServiceContainer,
	catchAllHandler chevron.Handler,
) (nacelle.Process, error) {
	setupRoutes := func(config nacelle.Config, router chevron.Router) error {
		router.AddMiddleware(middleware.NewLogging())
		router.AddMiddleware(NewControlMiddleware(catchAllHandler))

		router.MustRegister("/clear", &ClearResource{})
		router.MustRegister("/register", &RegisterResource{}, makeSchemaMiddleware())
		router.MustRegister("/requests", &RequestsResource{})
		router.MustRegister("/sse", &SSEResource{})
		return nil
	}

	server := basehttp.NewServer(chevron.NewInitializer(
		chevron.RouteInitializerFunc(setupRoutes),
		chevron.WithNotFoundHandler(catchAllHandler),
	))

	if err := services.Inject(server); err != nil {
		return nil, err
	}

	if err := server.Init(config); err != nil {
		return nil, err
	}

	return server, nil
}

func makeSchemaMiddleware() chevron.MiddlewareConfigFunc {
	schemaOptions := []middleware.SchemaConfigFunc{
		middleware.WithSchemaUnprocessableEntityFactory(unprocessableEntityFactory),
	}

	return chevron.WithMiddlewareFor(
		middleware.NewSchemaMiddleware("./schemas/handler.yaml", schemaOptions...),
		chevron.MethodPost,
	)
}

func unprocessableEntityFactory(errors []gojsonschema.ResultError) response.Response {
	details := map[string]string{}
	for _, err := range errors {
		details[err.Field()] = err.Description()
	}

	resp := response.JSON(map[string]interface{}{
		"error": details,
	})

	resp.SetStatusCode(http.StatusUnprocessableEntity)
	return resp
}
