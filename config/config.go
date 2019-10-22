package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
)

var config map[string]interface{}

var configValidation = map[string]reflect.Kind{
	"appName":       reflect.String,
	"environment":   reflect.String,
	"host":          reflect.String,
	"port":          reflect.Float64,
	"protocol":      reflect.String,
	"strictSlash":   reflect.Bool,
	"debug":         reflect.Bool,
	"timeout":       reflect.Float64,
	"dbConnection":  reflect.String,
	"dbHost":        reflect.String,
	"dbPort":        reflect.Float64,
	"dbName":        reflect.String,
	"dbUsername":    reflect.String,
	"dbPassword":    reflect.String,
	"dbOptions":     reflect.String,
	"dbAutoMigrate": reflect.Bool,
}

var authorizedValues = map[string][]string{
	"protocol":     []string{"http", "https"},
	"dbConnection": []string{"mysql", "postgres", "sqlite3", "mssql"},
}

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

	validateConfig()

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

func inSlice(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func validateConfig() {
	valid := true
	for key, value := range config {
		if v, ok := configValidation[key]; ok {
			t := reflect.TypeOf(value)
			if t.Kind() != v {
				log.Println(fmt.Sprintf("Invalid config entry. %s type must be %s", key, v))
				valid = false
				continue
			}

			if v, ok := authorizedValues[key]; ok {
				if !inSlice(v, value.(string)) {
					log.Println(fmt.Sprintf("Invalid config entry. %s must have one of the following values: %s", key, strings.Join(v, ", ")))
					valid = false
				}
			}
		}
	}

	if !valid {
		os.Exit(1)
	}
}
