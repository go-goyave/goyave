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
	Value            interface{}
	Type             reflect.Kind
	AuthorizedValues []interface{} // Leave empty for "any"
}

// TODO default values are stored in another map, nil if it's not required

var config object

var configDefaults object = object{
	"app": object{
		"name":            &Entry{"goyave", reflect.String, []interface{}{}},
		"environment":     &Entry{"localhost", reflect.String, []interface{}{}},
		"debug":           &Entry{true, reflect.Bool, []interface{}{}},
		"defaultLanguage": &Entry{"en-US", reflect.String, []interface{}{}},
	},
	"server": object{
		"host":          &Entry{"127.0.0.1", reflect.String, []interface{}{}},
		"domain":        &Entry{"", reflect.String, []interface{}{}},
		"protocol":      &Entry{"http", reflect.String, []interface{}{"http", "https"}},
		"port":          &Entry{8080, reflect.Int, []interface{}{}},
		"httpsPort":     &Entry{8081, reflect.Int, []interface{}{}},
		"timeout":       &Entry{10, reflect.Int, []interface{}{}},
		"maxUploadSize": &Entry{10, reflect.Int, []interface{}{}},
		"maintenance":   &Entry{false, reflect.Bool, []interface{}{}},
		"tlsCert":       &Entry{nil, reflect.String, []interface{}{}},
		"tlsKey":        &Entry{nil, reflect.String, []interface{}{}},
	},
	"database": object{
		"connection":         &Entry{"none", reflect.String, []interface{}{"none", "mysql", "postgres", "sqlite3", "mssql"}}, // TODO add a dialect ?
		"host":               &Entry{"127.0.0.1", reflect.String, []interface{}{}},
		"port":               &Entry{3306, reflect.Int, []interface{}{}},
		"name":               &Entry{"goyave", reflect.String, []interface{}{}},
		"username":           &Entry{"root", reflect.String, []interface{}{}},
		"password":           &Entry{"root", reflect.String, []interface{}{}},
		"options":            &Entry{"charset=utf8&parseTime=true&loc=Local", reflect.String, []interface{}{}},
		"maxOpenConnections": &Entry{20, reflect.Int, []interface{}{}},
		"maxIdleConnections": &Entry{20, reflect.Int, []interface{}{}},
		"maxLifetime":        &Entry{300, reflect.Int, []interface{}{}},
		"autoMigrate":        &Entry{false, reflect.Bool, []interface{}{}},
	},
	"jwt": object{ // TODO move this config to auth package
		"expiry": &Entry{300, reflect.Int, []interface{}{}},
		"secret": &Entry{nil, reflect.String, []interface{}{}},
	},
	// TODO don't forget optional entries in docs (such as tlsCert)
}

// TODO implement Register()

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

	if err := config.validate(""); err != nil {
		return fmt.Errorf("Invalid config:%s", err.Error())
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
			// TODO test if we're at the end of the path
			val := entry.(*Entry).Value
			return val, val != nil // nil means unset
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

// Set a config entry.
// The change is temporary and will not be saved for next boot.
// Use "nil" to unset a value.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	currentCategory := config
	path := strings.Split(key, ".")
	for _, catKey := range path {
		entry, ok := currentCategory[catKey]
		if !ok {
			// TODO document this behavior
			// If entry doesn't exist (and is not registered),
			// register it with the type of the type given here
			// and "any" authorized values.
			currentCategory[catKey] = &Entry{value, reflect.TypeOf(value).Kind(), []interface{}{}}
			return
		}

		if category, ok := entry.(object); ok {
			currentCategory = category
		} else {
			// TODO test if we're at the end of the path
			entry := currentCategory[catKey].(*Entry)
			entry.Value = value
			if err := entry.validate(key); err != nil {
				panic(err)
			}
			currentCategory[catKey] = entry
			return
		}
	}
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
				return fmt.Errorf("Invalid config:\n\t- Cannot override entry %q because it is a category", k)
			}
			override(obj, dst[k].(object))
		} else if entry, ok := dst[k]; ok {
			entry.(*Entry).Value = v
		} else {
			// TODO document this behavior
			// If entry doesn't exist (and is not registered),
			// register it with the type of the type given here
			// and "any" authorized values.
			dst[k] = &Entry{v, reflect.TypeOf(v).Kind(), []interface{}{}}
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

func (o object) validate(key string) error {
	message := ""
	valid := true
	for k, entry := range o {
		var subKey string
		if key == "" {
			subKey = k
		} else {
			subKey = key + "." + k
		}
		if category, ok := entry.(object); ok {
			if err := category.validate(subKey); err != nil {
				message += err.Error()
				valid = false
			}
		} else if err := entry.(*Entry).validate(subKey); err != nil {
			message += "\n\t- " + err.Error()
			valid = false
		}
	}

	if !valid {
		return fmt.Errorf(message)
	}
	return nil
}

func (e *Entry) validate(key string) error {
	if e.Value == nil { // nil values means unset
		return nil
	}

	kind := reflect.TypeOf(e.Value).Kind()
	if kind != e.Type {
		if !e.tryIntConversion(kind) {
			return fmt.Errorf("%q type must be %s", key, e.Type)
		}
		return nil
	}

	if len(e.AuthorizedValues) > 0 && !helper.Contains(e.AuthorizedValues, e.Value) {
		return fmt.Errorf("%q must have one of the following values: %v", key, e.AuthorizedValues)
	}

	return nil
}

func (e *Entry) tryIntConversion(kind reflect.Kind) bool {
	if kind == reflect.Float64 && e.Type == reflect.Int {
		intVal := int(e.Value.(float64))
		if e.Value == float64(intVal) {
			e.Value = intVal
			return true
		}
	}

	return false
}
