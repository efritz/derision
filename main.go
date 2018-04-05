package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/efritz/response"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

const (
	defaultRequestLogCapacity = 100
	negroniLogTemplate        = `{{.Method}} {{.Path}} -> {{.Status}}`
)

var requestLogCapacity = defaultRequestLogCapacity

func main() {
	if err := parseRequestLogCapacity(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not parse request log capacity\n")
		os.Exit(1)
	}

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

func parseRequestLogCapacity() error {
	if raw, ok := os.LookupEnv("REQUEST_LOG_CAPACITY"); ok && raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}

		requestLogCapacity = val
	}

	return nil
}
