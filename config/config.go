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
	"appName":              reflect.String,
	"environment":          reflect.String,
	"host":                 reflect.String,
	"port":                 reflect.Float64,
	"protocol":             reflect.String,
	"debug":                reflect.Bool,
	"timeout":              reflect.Float64,
	"maxUploadSize":        reflect.Float64,
	"dbConnection":         reflect.String,
	"dbHost":               reflect.String,
	"dbPort":               reflect.Float64,
	"dbName":               reflect.String,
	"dbUsername":           reflect.String,
	"dbPassword":           reflect.String,
	"dbOptions":            reflect.String,
	"dbMaxOpenConnections": reflect.Float64,
	"dbMaxIdleConnections": reflect.Float64,
	"dbAutoMigrate":        reflect.Bool,
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

	log.Panicf("Config entry %s doesn't exist", key)
	return nil
}

// GetString a config entry as string
func GetString(key string) string {
	val, ok := config[key]
	if ok {
		str, ok := val.(string)
		if !ok {
			log.Panicf("Config entry %s is not a string", key)
		}
		return str
	}

	log.Panicf("Config entry %s doesn't exist", key)
	return ""
}

// GetBool a config entry as bool
func GetBool(key string) bool {
	val, ok := config[key]
	if ok {
		b, ok := val.(bool)
		if !ok {
			log.Panicf("Config entry %s is not a bool", key)
		}
		return b
	}

	log.Panicf("Config entry %s doesn't exist", key)
	return false
}

func loadDefaults() error {
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		confDefaults, err := readConfigFile(path.Dir(filename) + string(os.PathSeparator) + "defaults.json")

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
