package main

import (
	log "github.com/sirupsen/logrus"
)
var (
	Environment Env
	Config      Conf
	Projects    []Project
)

func init() {
	log.SetLevel(log.TraceLevel)
	Environment = initEnvironment()
	initConfig()
	initLogging()
	initProjects()
}

func main() {

}
