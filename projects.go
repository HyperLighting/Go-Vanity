package main

import (
	"encoding/json"
	"errors"
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

// vanityURL builds the url to a specific project
func (project Project) vanityURL() string {
	return Config.Server.FQDomain() + "/" + project.ProjectPath
}

// docsURL works through several options to find a suitable url to use for the documentation. This works through
// static options, including checking that the folder is valid before serving that option, the enabled option
// and then all other documentation links. Finally if all else fails it will fall back to the main directory URL.
func (project Project) docsURL() string {
	// Try what ever is enabled
	if project.isServingStaticDocs() {
		if project.validStaticFolder() {
			return project.vanityURL()
		}
	} else {
		// Check the enabled Docs is valid
		if url, ok := project.Docs[project.EnabledDocs]; ok {
			return url
		}
	}

	// Not static, and no valid docs url, try other docs urls
	if len(project.Docs) > 0 {
		for k, url := range project.Docs {
			if k != "static" && k != "Static" && k != "STATIC" {
				return url
			}
		}
	}

	// No docs set, try the repo
	repo, err := project.getRepo()

	if err != nil {
		log.Error(err)
	} else {
		return repo.URL
	}

	// No repo, fall back to main directory
	return Config.Server.FQDomain()
}

// getEnabledRepo returns a valid repo info. This will first try to use the enabled repo, if it is valid. If it is
// not valid, it will loop through all of the repos in the project, and return the first it finds that is valid.
// If no valid repos can be found, it will return an error.
func (project Project) getRepo() (repo Repo, err error) {
	// Try the enabled project
	if en, ok := project.Repos[project.EnabledRepo]; ok {
		if valid, _ := en.isValid(); valid {
			return en, nil
		}
	}

	// Try other projects or fail
	if len(project.Repos) > 0 {
		// Try all the repos that are registered
		for _, r := range project.Repos {
			if valid, _ := r.isValid(); valid {
				log.Error("repo not found, using alternate")
				return r, nil
			}
		}

		// None of the repos are valid, return an error and fail
		return Repo{}, errors.New("repo not found, no valid repos found")
	} else {
		return Repo{}, errors.New("no repos found")
	}
}

// isValid check is the provided repo has a URL and Type which are the minimum pieces of data required for other parts of
// the system.
func (repo Repo) isValid() (valid bool, err error) {
	// Check the URL is set
	if repo.URL == "" {
		return false, errors.New("repo not valid, missing URL")
	}

	// Check the repo type is set
	if repo.Type == "" {
		return false, errors.New("repo not valid, missing Type")
	}

	// Default to true as it contains the required info
	return true, nil
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
	if project.isServingStaticDocs() && project.validStaticFolder() {
		// Serve the static folder
		folder := Config.Server.StaticDir + "/" + project.Docs[project.EnabledDocs]
		serveStaticFolder(folder, true, "/"+project.ProjectPath+"/").ServeHTTP(w, r)
	}

	// Redirect to the docs URL
	url := project.docsURL()
	redirect(w, r, url)
}
