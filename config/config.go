package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

var config map[string]interface{}

// LoadConfig loads the config.json file in the current working directory.
// If the "GOYAVE_ENV" env variable is set, the config file will be picked like so:
// - "production": "config.production.json"
// - "test": "config.test.json"
// - By default: "config.json"
func LoadConfig() error {
	err := loadDefaults()
	if err == nil {
		workingDir, err := os.Getwd()
		if err == nil {
			path := getConfigFilePath()
			conf, err := readConfigFile(fmt.Sprintf("%s%s%s", workingDir, string(os.PathSeparator), path))
			if err == nil {
				for key, value := range conf {
					config[key] = value
				}
			}
		}
	}

	return err
}

// Get a config entry
func Get(key string) interface{} {
	val, ok := config[key]
	if ok {
		return val
	}
	return nil
}

func loadDefaults() error {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		confDefaults, err := readConfigFile(fmt.Sprintf("%s%s%s", path.Dir(filename), string(os.PathSeparator), "defaults.json"))

		if err == nil {
			config = confDefaults
		}
		return err
	}
	return fmt.Errorf("Runtime caller error")
}

func readConfigFile(file string) (map[string]interface{}, error) {
	conf := map[string]interface{}{}
	configFile, err := os.Open(file)
	defer configFile.Close()

	if err == nil {
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(&conf)
	}

	return conf, err
}

func getConfigFilePath() string {
	switch strings.ToLower(os.Getenv("GOYAVE_ENV")) {
	case "test":
		return "config.test.json"
	case "production":
		return "config.production.json"
	default:
		return "config.json"
	}
}
