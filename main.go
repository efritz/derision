package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func main() {
	negroni := negroni.Classic()
	negroni.UseHandler(makeRouter(newServer()))
	negroni.Run("0.0.0.0:5000")
}

func makeRouter(s *server) *mux.Router {
	r := mux.NewRouter().StrictSlash(false)
	r.PathPrefix("/_control").Path("/clear").Methods("POST").HandlerFunc(s.clearHandler)
	r.PathPrefix("/_control").Path("/register").Methods("POST").HandlerFunc(s.registerHandler)
	r.PathPrefix("/_control").Path("/gather").Methods("GET").HandlerFunc(s.gatherHandler)
	r.NotFoundHandler = http.HandlerFunc(s.apiHandler)

	return r
}
