package main

import (
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var (
	Server *http.Server
	Mux    *http.ServeMux
)

func startServer() {
	// Build the server
	Server = buildServer()

	// Start the new server
	log.Fatal(Server.ListenAndServe())
}

func buildServer() (server *http.Server) {
	// Build the Mux
	buildMux()

	// Set up the Server
	server = &http.Server{
		Addr:    ":" + strconv.Itoa(Config.Server.Port),
		Handler: Mux,
	}

	return server
}

func buildMux() {
	// Build the mux
	Mux = http.NewServeMux()

	// Register the directory
	Mux.Handle("/", http.FileServer(http.Dir("./directory")))

	// Register projects
	handleProjects(Mux)
}
