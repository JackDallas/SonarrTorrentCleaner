package main

import (
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	ErrInvalidConfigFile      = errors.New("invalid Config File")
	ErrFailedToFindConfigFile = errors.New("failed to find config file")
)

type Config struct {
	CheckTimeMinutes         float64 `yaml:"CheckTimeMinutes"`
	NoProgressTimeoutMinutes float64 `yaml:"NoProgressTimeoutMinutes"`
	SonarrURL                string  `yaml:"SonarrURL"`
	SonarrAPIKey             string  `yaml:"SonarrAPIKey"`
	Blacklist                bool    `yaml:"Blacklist"`
	WebServer                bool    `yaml:"WebServer"`
	IgnoreQueued             bool    `yaml:"IgnoreQueued"`
	WebRoot                  string  `yaml:"WebRoot"`
	BindIP                   string  `yaml:"BindIP"`
	BindPort                 string  `yaml:"BindPort"`
}

func loadConfigFromDisk() (Config, error) {
	var config Config
	file, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		return config, ErrFailedToFindConfigFile
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return config, ErrInvalidConfigFile
	}

	return config, nil
}

func createDefaultConfig() error {
	config := Config{
		CheckTimeMinutes:         10,
		NoProgressTimeoutMinutes: 30,
		SonarrURL:                "http://localhost:8989",
		SonarrAPIKey:             "",
		Blacklist:                true,
		WebServer:                true,
		IgnoreQueued:             true,
		WebRoot:                  "",
		BindIP:                   "127.0.0.1",
		BindPort:                 "5252",
	}

	file, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("config.yaml", file, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadOrCreateConfig() (Config, error) {
	config, err := loadConfigFromDisk()
	if err != nil {
		if err == ErrFailedToFindConfigFile {
			err = createDefaultConfig()
			if err != nil {
				return config, err
			}
			panic("Default config created, please fill it out")
		}
		if err == ErrInvalidConfigFile {
			return config, ErrInvalidConfigFile
		}
	}
	//Clean up url
	if strings.HasSuffix(config.SonarrURL, ("/")) {
		config.SonarrURL = config.SonarrURL[:len(config.SonarrURL)-1]
	}

	return config, nil
}
