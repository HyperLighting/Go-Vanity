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

func (server ServerConfig) FQDomain() string {
	if server.UseSSL {
		return "https://" + server.Hostname
	} else {
		return "http://" + server.Hostname
	}
}

func serveStaticFolder(directory string, strip bool, stripPrefix string) http.Handler {
	log.WithFields(log.Fields{
		"Directory":   directory,
		"Strip":       strip,
		"StripPrefix": stripPrefix,
	}).Debug("Serving Stating Folder")

	if folderExists(directory) {
		fs := http.FileServer(http.Dir(directory))
		if strip {
			return http.StripPrefix(stripPrefix, fs)
		}
		return fs
	}
	log.Error("Folder doesn't exist, falling back to redirect to home")
	return http.RedirectHandler(Config.Server.FQDomain(), http.StatusFound)
}
