package main

import (
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
		FileName    string `yaml:"FileName" json:"FileName" env:"FILENAME" env-default:"projects.json"`
		MetaRefresh struct {
			Enabled bool   `yaml:"Enabled" json:"Enabled" env:"ENABLED" env-default:"false"`
			To      string `yaml:"To" json:"To" env:"TO" env-default:"docs"`
		} `yaml:"MetaRefresh" json:"MetaRefresh" env-prefix:"METAREFRESH_"`
		Remote struct {
			Enabled          bool   `yaml:"Enabled" json:"Enabled" env:"ENABLED" env-default:"false"`
			RefreshFrequency string `yaml:"RefreshFrequency" json:"RefreshFrequency" env:"REFRESH" env-default:"0 0 0 0 0 0"`
			URL              string `yaml:"URL" json:"URL" env:"URL"`
		} `yaml:"Remote" json:"Remote" env-prefix:"REMOTE_"`
	} `yaml:"Projects" json:"Projects" env-prefix:"VANITY_PROJECTS_"`
	Logging struct {
		Method string `yaml:"Method" json:"Method" env:"METHOD" env-default:"stdout"`
		Format string `yaml:"Format" json:"Format" env:"FORMAT" env-default:"text"`
		Level  string `yaml:"Level" json:"Level" env:"LEVEL" env-default:"Error"`
		File   string `yaml:"File" json:"File" env:"FILE" env-default:"vanity.log"`
	} `yaml:"Logging" json:"Logging" env-prefix:"VANITY_LOGGING_"`
}

func initConfig() {
	// Try to load a config file?
	var configFile bool = true
	var configFileName string = "config.yaml"
	noConfFile, ok := Environment["NO_CONFIG_FILE"]

	if ok {
		noConfFileb, err := strconv.ParseBool(noConfFile.(string))

		if err != nil {
			log.WithFields(log.Fields{
				"No Config File ENV Value": noConfFile,
			}).Error(err)
		}

		if !noConfFileb {
			configFile = false
		}
	}

	// Should we try a different config file name?
	if name, ok := Environment["CONFIG_FILE"]; ok {
		configFileName = name.(string)
	}

	// Load the config
	if configFile {
		log.WithFields(log.Fields{
			"Config File": configFileName,
		}).Debug("Reading Config File")

		cleanenv.ReadConfig(configFileName, &Config)
		return
	} else {
		log.Debug("No Config File, reading ENV Variables")
		cleanenv.ReadEnv(&Config)
		return
	}

}
