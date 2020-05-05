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

var config map[string]interface{}

var configDefaults map[string]interface{} = map[string]interface{}{
	"appName":              "goyave",
	"environment":          "localhost",
	"maintenance":          false,
	"host":                 "127.0.0.1",
	"domain":               "",
	"port":                 8080.0,
	"httpsPort":            8081.0,
	"protocol":             "http",
	"debug":                true,
	"timeout":              10.0,
	"maxUploadSize":        10.0,
	"defaultLanguage":      "en-US",
	"dbConnection":         "none",
	"dbHost":               "127.0.0.1",
	"dbPort":               3306.0,
	"dbName":               "goyave",
	"dbUsername":           "root",
	"dbPassword":           "root",
	"dbOptions":            "charset=utf8&parseTime=true&loc=Local",
	"dbMaxOpenConnections": 20.0,
	"dbMaxIdleConnections": 20.0,
	"dbMaxLifetime":        300.0,
	"dbAutoMigrate":        false,
	"jwtExpiry":            300.0,
}

var configValidation = map[string]reflect.Kind{
	"appName":              reflect.String,
	"environment":          reflect.String,
	"maintenance":          reflect.Bool,
	"host":                 reflect.String,
	"domain":               reflect.String,
	"port":                 reflect.Float64,
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
	"dbConnection": {"none", "mysql", "postgres", "sqlite3", "mssql"},
}
var mutex = &sync.RWMutex{}

// Load loads the config.json file in the current working directory.
// If the "GOYAVE_ENV" env variable is set, the config file will be picked like so:
// - "production": "config.production.json"
// - "test": "config.test.json"
// - By default: "config.json"
func Load() error {
	mutex.Lock()
	defer mutex.Unlock()
	loadDefaults()
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := getConfigFilePath()
	conf, err := readConfigFile(fmt.Sprintf("%s%s%s", workingDir, string(os.PathSeparator), path))
	if err != nil {
		return err
	}

	for key, value := range conf {
		config[key] = value
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

// Get a config entry
func Get(key string) interface{} {
	mutex.RLock()
	val, ok := config[key]
	mutex.RUnlock()
	if ok {
		return val
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

// Has check if a config entry exists.
func Has(key string) bool {
	mutex.RLock()
	_, ok := config[key]
	mutex.RUnlock()
	return ok
}

// Set a config entry
//
// The change is temporary and will not be saved for next boot.
func Set(key string, value interface{}) {
	if err := validateEntry(value, key); err != nil {
		panic(err)
	}
	mutex.Lock()
	config[key] = value
	mutex.Unlock()
}

// GetString a config entry as string
func GetString(key string) string {
	mutex.RLock()
	val, ok := config[key]
	mutex.RUnlock()
	if ok {
		str, ok := val.(string)
		if !ok {
			panic(fmt.Sprintf("Config entry \"%s\" is not a string", key))
		}
		return str
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

// GetBool a config entry as bool
func GetBool(key string) bool {
	mutex.RLock()
	val, ok := config[key]
	mutex.RUnlock()
	if ok {
		b, ok := val.(bool)
		if !ok {
			panic(fmt.Sprintf("Config entry \"%s\" is not a bool", key))
		}
		return b
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

// Register a config entry for validation.
// If the entry identified by the given key is set or modified, its
// value will be validated according to the given kind.
// If the entry already exists, it will be revalidated.
//
// This method doesn't allow to override existing validation. Once an
// entry is registered, its expected kind cannot be modified.
func Register(key string, kind reflect.Kind) {
	_, exists := configValidation[key]
	if exists {
		panic(fmt.Sprintf("Config entry \"%s\" is already registered", key))
	}

	configValidation[key] = kind

	val, exists := config[key]
	if exists {
		if err := validateEntry(val, key); err != nil {
			delete(configValidation, "key")
			panic(err)
		}
	}

}

func loadDefaults() {
	config = make(map[string]interface{}, len(configDefaults))
	for k, v := range configDefaults {
		config[k] = v
	}
}

func readConfigFile(file string) (map[string]interface{}, error) {
	conf := map[string]interface{}{}
	configFile, err := os.Open(file)

	if err == nil {
		defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&conf)
	}

	return conf, err
}

func getConfigFilePath() string {
	env := strings.ToLower(os.Getenv("GOYAVE_ENV"))
	if env == "local" || env == "localhost" || env == "" {
		return "config.json"
	}
	return "config." + env + ".json"
}

func validateConfig() error {
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

func validateEntry(value interface{}, key string) error {
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
