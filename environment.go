package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	SystemEnvironment string = "Production"
)

type Env map[string]interface{}

func initEnvironment() Env {
	env := make(Env)
	envVariables := [...]string{
		"ENVIRONMENT",
		"NO_CONFIG_FILE",
		"CONFIG_FILE",
		"PROJECTS_FILE",
		"PORT",
		"STATIC_DIR",
		"BASEURL",
		"USESSL",
		"METAREFRESH",
		"METAREFRESH_TO",
		"LOG_METHOD",
		"LOG_FORMAT",
		"LOG_LEVEL",
		"LOG_FILE",
		"FETCH_PROJECT_ENABLED",
		"FETCH_PROJECT_URL",
		"FETCH_PROJECT_FREQUENCY",
	}

	// Loop through all variables and build a map for use later
	for _, e := range envVariables {
		// All Environment Variables are Pre-Fixed with Vanity
		log.Trace("Checking Environment Variable: " + e)
		if val, present := os.LookupEnv("VANITY_" + e); present {
			env[e] = val
			log.WithFields(log.Fields{
				"Variable": e,
				"Value":    val,
			}).Trace("Environment Variable Found")
		}
	}

	// Set the System Environment if that has been set
	if val, present := env["ENVIRONMENT"]; present {
		SystemEnvironment = val.(string)
	}

	return env
}
