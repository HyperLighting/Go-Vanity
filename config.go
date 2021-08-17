package main

import (
	"os"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
	log "github.com/sirupsen/logrus"
)

type Conf struct {
	Server struct {
		Port      int    `yaml:"Port" json:"Port" env:"PORT" env-default:"8080"`
		Hostname  string `yaml:"Hostname" json:"Hostname" env:"HOSTNAME" env-default:"go.example.dev"`
		UseSSL    bool   `yaml:"UseSSL" json:"UseSSL" env:"USESSL" env-default:"true"`
		StaticDir string `yaml:"StaticDir" json:"StaticDir" env:"STATICDIR" env-default:"static"`
	} `yaml:"Server" json:"Server" env-prefix:"VANITY_SERVER_"`
	Projects struct {
		Source       string `yaml:"Source" json:"Source" env:"SOURCE" env-default:"projects.yml"`
		SourceType   string `yaml:"SourceType" json:"SourceType" env:"SOURCETYPE" env-default:"local"`
		SourceFormat string `yaml:"SourceFormat" json:"SourceFormat" env:"SOURCEFORMAT" env-default:"yaml"`
		Refresh      struct {
			Enabled   bool   `yaml:"Enabled" json:"Enabled" env:"REFRESH" env-default:"false"`
			Frequency string `yaml:"Frequency" json:"Frequency" env:"FREQUENCY" env-default:"0 0 * * *"`
		} `yaml:"Refresh" json:"Refresh" env-prefix:"REFRESH_"`
		MetaRefresh struct {
			Enabled bool   `yaml:"Enabled" json:"Enabled" env:"ENABLED" env-default:"true"`
			To      string `yaml:"To" json:"To" env:"TO" env-default:"repo"`
		} `yaml:"MetaRefresh" json:"MetaRefresh" env-prefix:"METAREFRESH"`
	} `yaml:"Projects" json:"Projects" env-prefix:"VANITY_PROJECTS_"`
	Logging struct {
		Default struct {
			Method string `yaml:"Method" json:"Method" env:"METHOD" env-default:"stdout"`
			Format string `yaml:"Format" json:"Format" env:"FORMAT" env-default:"text"`
			Level  string `yaml:"Level" json:"Level" env:"LEVEL" env-default:"Error"`
			File   string `yaml:"File" json:"File" env:"FILE" env-default:"vanity.log"`
		} `yaml:"Default" json:"Default" env-prefix:"DEFAULT_"`
		Prod struct {
			Method string `yaml:"Method" json:"Method" env:"METHOD" env-default:"stderr"`
			Format string `yaml:"Format" json:"Format" env:"FORMAT" env-default:"text"`
			Level  string `yaml:"Level" json:"Level" env:"LEVEL" env-default:"Error"`
			File   string `yaml:"File" json:"File" env:"FILE" env-default:"vanity.log"`
		} `yaml:"Prod" json:"Prod" env-prefix:"PROD_"`
		Dev struct {
			Method string `yaml:"Method" json:"Method" env:"METHOD" env-default:"file"`
			Format string `yaml:"Format" json:"Format" env:"FORMAT" env-default:"text"`
			Level  string `yaml:"Level" json:"Level" env:"LEVEL" env-default:"Debug"`
			File   string `yaml:"File" json:"File" env:"FILE" env-default:"vanity.log"`
		} `yaml:"Dev" json:"Dev" env-prefix:"DEV_"`
	} `yaml:"Logging" json:"Logging" env-prefix:"VANITY_LOGGING_"`
}

func initConfig() {
	// Try to load a config file?
	var configFile bool = true
	var configFileName string = "config.yaml"
	noConfFile, ok := Environment["NO_CONFIG_FILE"]

	if ok {
		log.Trace("No Config File Env Variable Found")
		noConfFileb, err := strconv.ParseBool(noConfFile.(string))

		if err != nil {
			log.WithFields(log.Fields{
				"No Config File ENV Value": noConfFile,
			}).Error(err)
		}

		if !noConfFileb {
			log.Trace("No Config File Env Variable has disabled loading the config file")
			configFile = false
		}
	}

	// Should we try a different config file name?
	if name, ok := Environment["CONFIG_FILE"]; ok {
		log.WithFields(log.Fields{
			"File Name": name.(string),
		}).Debug("Config File Env Variable has changed the name of the file we are looking for")
		configFileName = name.(string)
	}

	// Load the config
	if configFile {
		// Does the file exist?
		if _, err := os.Stat(configFileName); err == nil {
			// File Exists, Load it
			log.WithFields(log.Fields{
				"Config File": configFileName,
			}).Debug("Reading Config File")
			cleanenv.ReadConfig(configFileName, &Config)
			return
		} else if os.IsNotExist(err) {
			// File doesn't exist, default to try ENV
			log.WithFields(log.Fields{
				"File Name": configFileName,
			}).Error("Config File Missing, Defaulting to ENV")
			cleanenv.ReadEnv(&Config)
			return
		} else {
			// Error, but not that the file doesn't exist. Log the error, and default to ENV
			log.WithFields(log.Fields{
				"File Name": configFileName,
			}).Error(err)
			cleanenv.ReadEnv(&Config)
			return
		}
	} else {
		log.Debug("No Config File, reading ENV Variables")
		cleanenv.ReadEnv(&Config)
		return
	}
}
