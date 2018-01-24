package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/efritz/response"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

const negroniLogTemplate = `{{.Method}} {{.Path}} -> {{.Status}}`

var (
	registrationPath = kingpin.Flag("registration-path", "Path to file with _control/register payloads").String()
	maxRequestLog    = kingpin.Flag("max-request-log", "Maximum number of requests to hold in memory").Default("100").Int()
)

func main() {
	kingpin.Parse()

	if err := run(); err != nil {
		if verr, ok := err.(*validationError); ok {
			fmt.Fprintf(os.Stderr, "error: failed to validate registration file:\n")

			for _, err := range verr.errors {
				fmt.Fprintf(os.Stderr, "    - %s\n", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		}

		os.Exit(1)
	}
}

func run() error {
	handlers, err := makeHandlersFromFile(*registrationPath)
	if err != nil {
		return err
	}

	server := newServer(*maxRequestLog)
	for _, handler := range handlers {
		server.addHandler(handler)
	}

	// Setup logging
	logger := negroni.NewLogger()
	logger.SetFormat(negroniLogTemplate)

	// Start the server
	negroni := negroni.New(negroni.NewRecovery(), logger)
	negroni.UseHandler(makeRouter(server))
	negroni.Run("0.0.0.0:5000")
	return nil
}

func makeRouter(s *server) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	r.PathPrefix("/_control").Path("/register").Methods("POST").HandlerFunc(response.Convert(s.registerHandler))
	r.PathPrefix("/_control").Path("/clear").Methods("POST").HandlerFunc(response.Convert(s.clearHandler))
	r.PathPrefix("/_control").Path("/requests").Methods("GET").HandlerFunc(response.Convert(s.requestsHandler))
	r.NotFoundHandler = http.HandlerFunc(response.Convert(s.apiHandler))

	return r
}
