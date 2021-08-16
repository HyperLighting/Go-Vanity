package main

import (
	log "github.com/sirupsen/logrus"
)
var (
	Environment Env
	Config      Conf
)

func init() {
	log.SetLevel(log.TraceLevel)
	Environment = initEnvironment()
	initConfig()
	initLogging()
}

func main() {

}
