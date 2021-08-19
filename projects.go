package main

import (
	"encoding/json"
	"net/http"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	initialProjectsLoadCompleted bool = false
)

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Project Types
// ------------------------------------------------------------------------------------------------------------------------------------------------------

type Project struct {
	Name        string            `yaml:"Name" json:"Name"`
	ProjectPath string            `yaml:"ProjectPath" json:"ProjectPath"`
	EnabledRepo string            `yaml:"EnabledRepo" json:"EnabledRepo"`
	EnabledDocs string            `yaml:"EnabledDocs" json:"EnabledDocs"`
	Repos       map[string]Repo   `yaml:"Repos" json:"Repos"`
	Docs        map[string]string `yaml:"Docs" json:"Docs"`
}

type Repo struct {
	Name   string `yaml:"Name" json:"Name"`
	Type   string `yaml:"Type" json:"Type"`
	URL    string `yaml:"URL" json:"URL"`
	Source Source `yaml:"Source" json:"Source"`
}

type Source struct {
	HomeURL      string `yaml:"HomeURL" json:"HomeURL"`
	DirectoryURL string `yaml:"DirectoryURL" json:"DirectoryURL"`
	FileLineURL  string `yaml:"FileLineURL" json:"FileLineURL"`
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Projects Initialisation
// ------------------------------------------------------------------------------------------------------------------------------------------------------

// initProjects handles calling the initial loading of the projects, and scheduling them to be reloaded if the config
// is set to refresh. If refresh is enabled but a malformed frequency is set, it will default to daily.
func initProjects() {
	// Are we loading a remote?
	if Config.Projects.Refresh.Enabled {
		// Initial Load
		log.Debug("Loading initial projects")
		loadProjects()

		// Schedule Refresh
		log.Debug("Scheduling Projects Refresh")
		c := cron.New()

		// Add the Refresh
		cid, err := c.AddFunc(Config.Projects.Refresh.Frequency, loadProjects)

		if err != nil {
			log.Error(err)
			log.Info("Defaulting to daily projects refresh")
			c.Remove(cid)
			c.AddFunc("0 0 * * *", loadProjects)
		}

		c.Start()
		log.Info("Cron starter to update projects")
	} else {
		log.Debug("Loading Projects Once")
		loadProjects()
	}
}

// loadProjects performs the actual loading of the projects, either remote or local.
func loadProjects() {
	// Remote or Local?
	var b []byte
	var e error
	switch Config.Projects.SourceType {
	case "local", "Local", "LOCAL":
		b, e = readLocalFile(Config.Projects.Source)
	case "remote", "Remote", "REMOTE":
		b, e = readRemoteFile(Config.Projects.Source)
	default:
		handleProjectsError(log.Fields{
			"Source Type": Config.Projects.SourceType,
		}, "Unable to load projects, Source Type not recognised")
		return
	}

	if e != nil {
		handleProjectsError(log.Fields{
			"Source":      Config.Projects.Source,
			"Source Type": Config.Projects.SourceType,
		}, e)
		return
	}

	// Unmarshal bytes to a new projects config
	var newProjects []Project
	var convErr error
	switch Config.Projects.SourceFormat {
	case "json", "JSON":
		convErr = json.Unmarshal(b, &newProjects)
	case "yaml", "YAML", "yml", "YML":
		convErr = yaml.Unmarshal(b, &newProjects)
	default:
		handleProjectsError(log.Fields{
			"Source Format": Config.Projects.SourceFormat,
		}, "Unable to load projects, Source Format not recognised")
		return
	}

	if convErr != nil {
		handleProjectsError(log.Fields{
			"Source":        Config.Projects.Source,
			"Source Type":   Config.Projects.SourceType,
			"Source Format": Config.Projects.SourceFormat,
		}, convErr)
		return
	}

	// Put the new Projects into action!
	log.Info("Projects refreshed")
	Projects = newProjects

	// Refresh the server on new projects load
	if initialProjectsLoadCompleted {
		// Currently it is not possible to update the existing projects
		// TODO: Find a way to update or restart the server that's running when refreshing projects
		return
	}

	// Ensure initial projects load is set
	initialProjectsLoadCompleted = true
}

// handleProjectsError is a helper function for handling errors in the loadProjects function. If this is the first
// time loading projects, it will cause a fatal error which is logged with the fields passed. If we are reloading
// projects it will just cause an error log as we can continue using the old version of projects.
func handleProjectsError(Fields log.Fields, msg ...interface{}) {
	if initialProjectsLoadCompleted {
		log.WithFields(Fields).Error(msg...)
	} else {
		log.WithFields(Fields).Fatal(msg...)
	}
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Projects Helper Functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

// isServingStaticDocs takes a project and evaluates the enabled docs property to determine if we are serving
// local files or redirecting to an external source
func (project Project) isServingStaticDocs() (res bool) {
	switch project.EnabledDocs {
	case "static", "Static", "STATIC":
		return true
	default:
		return false
	}
}

// validStaticFolder checks if the folder exists, but only if we are serving a static folder.
func (project Project) validStaticFolder() bool {
	if project.isServingStaticDocs() {
		return folderExists(Config.Server.StaticDir + "/" + project.Docs[project.EnabledDocs])
	}
	return false
}

// ------------------------------------------------------------------------------------------------------------------------------------------------------
// Projects HTTP Functions
// ------------------------------------------------------------------------------------------------------------------------------------------------------

// Handle deals with registering a project with the supplied mux. It also builds the function for deciding to call
// GenerateVanityPage or calling ServeDocs
func (project Project) Handle(mux *http.ServeMux) {
	mux.HandleFunc("/"+project.ProjectPath+"/", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Request for: " + project.Name)

		// Serve the download meta data
		if isGoGetRequest(r) {
			project.GenerateVanityPage(w, r)
			return
		}

		// Redirect to docs
		project.ServeDocs(w, r)
	})
}

// ServeDocs handles calls for the documentation of a project. It handles the decision making between serving static
// documents or redirecting another source.
func (project Project) ServeDocs(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "https://google.com")
}
