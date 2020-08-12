package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/System-Glitch/goyave/v2/helper"
)

type object map[string]interface{}

type Entry struct {
	Key              string // Full key
	Value            interface{}
	Type             reflect.Kind
	AuthorizedValues []interface{} // Leave empty for "any"
}

// TODO default values are stored in another map, nil if it's not required

var config object

var configDefaults object = object{
	"app": object{
		"name":            "goyave",
		"environment":     "localhost",
		"debug":           true,
		"defaultLanguage": "en-US",
	},
	"server": object{
		"host":          "127.0.0.1",
		"domain":        "",
		"protocol":      "http",
		"port":          8080,
		"httpsPort":     8081,
		"timeout":       10,
		"maxUploadSize": 10,
		"maintenance":   false,
	},
	"database": object{
		"connection":         "none",
		"host":               "127.0.0.1",
		"port":               3306,
		"name":               "goyave",
		"username":           "root",
		"password":           "root",
		"options":            "charset=utf8&parseTime=true&loc=Local",
		"maxOpenConnections": 20,
		"maxIdleConnections": 20,
		"maxLifetime":        300,
		"autoMigrate":        false,
	},
	"jwtExpiry": 300.0, // TODO move jwtExpry to auth package init func
	// TODO don't forget optional entries (such as tlsCert)
}

// TODO implement RegisterDefault()

var configValidation = map[string]reflect.Kind{ // TODO is config validation really useful?
	"appName":              reflect.String,
	"environment":          reflect.String,
	"maintenance":          reflect.Bool,
	"host":                 reflect.String,
	"domain":               reflect.String,
	"port":                 reflect.Float64, // TODO if expected is int, but received is float and decimal is 0, cast
	"httpsPort":            reflect.Float64,
	"protocol":             reflect.String,
	"debug":                reflect.Bool,
	"timeout":              reflect.Float64,
	"maxUploadSize":        reflect.Float64,
	"defaultLanguage":      reflect.String,
	"tlsCert":              reflect.String,
	"tlsKey":               reflect.String,
	"dbConnection":         reflect.String,
	"dbHost":               reflect.String,
	"dbPort":               reflect.Float64,
	"dbName":               reflect.String,
	"dbUsername":           reflect.String,
	"dbPassword":           reflect.String,
	"dbOptions":            reflect.String,
	"dbMaxOpenConnections": reflect.Float64,
	"dbMaxIdleConnections": reflect.Float64,
	"dbMaxLifetime":        reflect.Float64,
	"dbAutoMigrate":        reflect.Bool,
	"jwtExpiry":            reflect.Float64,
}

var authorizedValues = map[string][]string{
	"protocol":     {"http", "https"},
	"dbConnection": {"none", "mysql", "postgres", "sqlite3", "mssql"}, // TODO how to add dialect?
}
var mutex = &sync.RWMutex{}

// Load loads the config.json file in the current working directory.
// If the "GOYAVE_ENV" env variable is set, the config file will be picked like so:
// - "production": "config.production.json"
// - "test": "config.test.json"
// - By default: "config.json"
func Load() error { // TODO allow loading from somewhere else
	mutex.Lock()
	defer mutex.Unlock()
	config = make(object, len(configDefaults))
	loadDefaults(configDefaults, config)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := getConfigFilePath()
	conf, err := readConfigFile(fmt.Sprintf("%s%s%s", workingDir, string(os.PathSeparator), path))
	if err != nil {
		return err
	}

	if err := override(conf, config); err != nil {
		// TODO test override error
		return err
	}

	if err := validateConfig(); err != nil {
		return err
	}

	return err
}

// IsLoaded returns true if the config have been loaded.
func IsLoaded() bool {
	mutex.RLock()
	defer mutex.RUnlock()
	return config != nil
}

// Clear unloads the config.
// DANGEROUS, should only be used for testing.
func Clear() {
	mutex.Lock()
	config = nil
	mutex.Unlock()
}

// Get a config entry.
func Get(key string) interface{} {
	if val, ok := get(key); ok {
		return val
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

func get(key string) (interface{}, bool) {
	// TODO getting a category should not be allowed
	// because it could be modified without passing
	// through the validation system
	mutex.RLock()
	defer mutex.RUnlock()
	currentCategory := config
	path := strings.Split(key, ".")
	for _, catKey := range path {
		entry, ok := currentCategory[catKey]
		if !ok {
			break
		}

		if category, ok := entry.(object); ok {
			currentCategory = category
		} else {
			return entry, true
		}
	}
	return nil, false
}

// GetString a config entry as string. Panics if entry is not a string.
func GetString(key string) string {
	str, ok := Get(key).(string)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a string", key))
	}
	return str
}

// GetBool a config entry as bool. Panics if entry is not a bool.
func GetBool(key string) bool {
	val, ok := Get(key).(bool)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a bool", key))
	}
	return val
}

// GetInt a config entry as int. Panics if entry is not an int.
func GetInt(key string) int {
	val, ok := Get(key).(int)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not an int", key))
	}
	return val
}

// GetFloat a config entry as float64. Panics if entry is not a float64.
func GetFloat(key string) float64 { // TODO update accessors docs
	val, ok := Get(key).(float64)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a float64", key))
	}
	return val
}

// Has check if a config entry exists.
func Has(key string) bool {
	_, ok := get(key)
	return ok
}

// Set a config entry
//
// The change is temporary and will not be saved for next boot.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	currentCategory := config
	path := strings.Split(key, ".")
	for _, catKey := range path {
		entry, ok := currentCategory[catKey]
		if !ok {
			break
		}

		if category, ok := entry.(object); ok {
			currentCategory = category
		} else {
			// TODO validate here
			// (all entries should be stored in a struct containing their types and authorized values)
			currentCategory[catKey] = value
		}
	}
	config[key] = value
}

func loadDefaults(src object, dst object) {
	for k, v := range src {
		if obj, ok := v.(object); ok {
			sub := make(object, len(obj))
			loadDefaults(obj, sub)
			dst[k] = sub
		} else {
			dst[k] = v
		}
	}
}

func override(src object, dst object) error { // TODO test override
	for k, v := range src {
		if obj, ok := v.(map[string]interface{}); ok {
			if dstObj, ok := dst[k]; !ok {
				dst[k] = make(object, len(obj))
			} else if _, ok := dstObj.(object); !ok {
				// Conflict: destination is not a category
				return fmt.Errorf("Invalid config:\n\t- Cannot override entry \"%s\" because it is category", k)
				// TODO find a way to retrieve full key
			}
			override(obj, dst[k].(object))
		} else {
			dst[k] = v
		}
	}
	return nil
}

func readConfigFile(file string) (object, error) {
	conf := make(object, len(configDefaults))
	configFile, err := os.Open(file)

	if err == nil {
		defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&conf)
		// TODO use interface or something to let users use
		// other file formats such as toml, provided said format
		// can unmarshal to map[string]interface{}
	}

	// TODO load environment variables
	// if variable not set, config loading error
	// if variable type is not string, try to convert
	return conf, err
}

func getConfigFilePath() string {
	env := strings.ToLower(os.Getenv("GOYAVE_ENV"))
	if env == "local" || env == "localhost" || env == "" {
		return "config.json"
	}
	return "config." + env + ".json"
}

func validateConfig() error { // TODO update validate config
	message := "Invalid config:"
	valid := true
	for key, value := range config {
		if err := validateEntry(value, key); err != nil {
			message += "\n\t- " + err.Error()
			valid = false
		}
	}

	if !valid {
		return fmt.Errorf(message)
	}
	return nil
}

func validateEntry(value interface{}, key string) error { // TODO handle multi-level
	if v, ok := configValidation[key]; ok {
		t := reflect.TypeOf(value)
		if t.Kind() != v {
			return fmt.Errorf("%q type must be %s", key, v)
		}

		if v, ok := authorizedValues[key]; ok {
			if !helper.ContainsStr(v, value.(string)) {
				return fmt.Errorf("%q must have one of the following values: %s", key, strings.Join(v, ", "))
			}
		}
	}
	return nil
}
