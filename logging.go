package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type logging struct {
	Method string
	Format string
	Level  string
	File   string
}

func getLoggingConf() (lc logging) {
	// Determing what configuration to use
	switch SystemEnvironment {
	case "prod", "Prod", "PROD", "production", "Production", "PRODUCTION":
		lc.Method = Config.Logging.Prod.Method
		lc.Format = Config.Logging.Prod.Format
		lc.Level = Config.Logging.Prod.Level
		lc.File = Config.Logging.Prod.File
		log.Debug("Logging set to Production mode")
	case "dev", "Dev", "DEV", "develop", "Develop", "DEVELOP", "development", "Development", "DEVELOPMENT":
		lc.Method = Config.Logging.Dev.Method
		lc.Format = Config.Logging.Dev.Format
		lc.Level = Config.Logging.Dev.Level
		lc.File = Config.Logging.Dev.File
		log.Debug("Logging set to Development mode")
	default:
		lc.Method = Config.Logging.Default.Method
		lc.Format = Config.Logging.Default.Format
		lc.Level = Config.Logging.Default.Level
		lc.File = Config.Logging.Default.File
		log.Debug("Logging set to Default mode")
	}

	return lc
}

func initLogging() {
	lc := getLoggingConf()

	// Logging Method
	switch lc.Method {
	case "stdout", "Stdout", "STDOUT":
		log.SetOutput(os.Stdout)
	case "file", "File", "FILE":
		logToFileInit(lc.File)
	default:
		// STDERR
		log.SetOutput(os.Stderr)
	}

	// Logging Format
	switch lc.Format {
	case "json", "JSON":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{})
	}

	// Set the Level
	switch lc.Level {
	case "trace", "Trace", "TRACE":
		log.SetLevel(log.TraceLevel)
	case "debug", "Debug", "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "info", "Info", "INFO":
		log.SetLevel(log.InfoLevel)
	case "warn", "Warn", "WARN", "warning", "Warning", "WARNING":
		log.SetLevel(log.WarnLevel)
	case "error", "Error", "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "fatal", "Fatal", "FATAL":
		log.SetLevel(log.FatalLevel)
	case "panic", "Panic", "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}
}

func logToFileInit(fileName string) {
	// If no filename set, default to stderr
	if fileName == "" {
		log.SetOutput(os.Stderr)
		log.Error("Logging Filename is empty, defaulted to stderr")
		return
	}

	// Open the file, defer closing
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		log.SetOutput(os.Stderr)
		log.Error(err)
		return
	}

	defer file.Close()

	// Set the output to the file we have opened
	log.SetOutput(file)
}
