package main

import (
	"net/http"

	"github.com/efritz/response"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

const negroniLogTemplate = `{{.Method}} {{.Path}} -> {{.Status}}`

func main() {
	logger := negroni.NewLogger()
	logger.SetFormat(negroniLogTemplate)

	negroni := negroni.New(negroni.NewRecovery(), logger)
	negroni.UseHandler(makeRouter(newServer()))
	negroni.Run("0.0.0.0:5000")
}

func makeRouter(s *server) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	r.PathPrefix("/_control").Path("/register").Methods("POST").HandlerFunc(response.Convert(s.registerHandler))
	r.PathPrefix("/_control").Path("/clear").Methods("POST").HandlerFunc(response.Convert(s.clearHandler))
	r.PathPrefix("/_control").Path("/requests").Methods("GET").HandlerFunc(response.Convert(s.requestsHandler))
	r.NotFoundHandler = http.HandlerFunc(response.Convert(s.apiHandler))

	return r
}
