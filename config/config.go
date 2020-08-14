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

// Entry is the internal reprensentation of a config entry.
// It contains the entry value, its expected type (for validation)
// and a slice of authorized values (for validation too). If this slice
// is empty, it means any value can be used, provided it is of the correct type.
type Entry struct {
	Value            interface{}
	Type             reflect.Kind
	AuthorizedValues []interface{} // Leave empty for "any"
}

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
}

var mutex = &sync.RWMutex{}

// Register a new config entry and its validation.
//
// Each module should register its config entries in an "init()"
// function, even if they don't have a default value, in order to
// ensure they will be validated.
// Each module should use its own category and use a name both expressive
// and unique to avoid collisions.
// For example, the "auth" package registers, among others, "auth.basic.username"
// and "auth.jwt.expiry", thus creating a category for its package, and two subcategories
// for its features.
//
// To register an entry without a default value (only specify how it
// will be validated), set "Entry.Value" to "nil".
//
// Panics if an entry already exists for this key and is not identical to the
// one passed as parameter of this function. On the other hand, if the entries
// are identical, no conflict is expected so the configuration is left in its
// current state.
func Register(key string, entry Entry) { // TODO test register
	if key == "" {
		panic("Empty key is not allowed")
	}

	mutex.Lock()
	defer mutex.Unlock()
	category, entryKey, exists := walk(configDefaults, key)
	if exists {
		if !reflect.DeepEqual(entry, category[entryKey].(*Entry)) {
			panic(fmt.Sprintf("Attempted to override registered config entry %q", key))
		}
	} else {
		category[entryKey] = &entry
	}
}

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
	mutex.RLock()
	defer mutex.RUnlock()
	currentCategory := config
	b := 0
	e := strings.Index(key, ".")
	if e == -1 {
		e = len(key)
	}
	for path := key[b:e]; ; path = key[b:e] {
		entry, ok := currentCategory[path]
		if !ok {
			break
		}

		if category, ok := entry.(object); ok {
			currentCategory = category
		} else {
			val := entry.(*Entry).Value
			return val, val != nil // nil means unset
		}

		if e+1 <= len(key) {
			b = e + 1
			newE := strings.Index(key[b:], ".")
			if newE == -1 {
				e = len(key)
			} else {
				e = newE + b
			}
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
//
//  - A category cannot be replaced with an entry.
//  - An entry cannot be replaced with a category.
//  - New categories can be created with they don't already exist.
//  - New entries can be created if they don't already exist. This new entry
//    will be subsequently validated using the type of its initial value and
//    have an empty slice as authorized values (meaning it can have any value of its type)
//
// Panics in case of error.
func Set(key string, value interface{}) {
	if key == "" {
		panic("Empty key is not allowed")
	}

	mutex.Lock()
	defer mutex.Unlock()
	category, entryKey, exists := walk(config, key)
	if exists {
		entry := category[entryKey].(*Entry)
		entry.Value = value
		if err := entry.validate(key); err != nil {
			panic(err)
		}
		category[entryKey] = entry
	} else {
		category[entryKey] = &Entry{value, reflect.TypeOf(value).Kind(), []interface{}{}}
	}
}

// walk the config using the key. Returns the deepest category, the entry key
// with its path stripped ("app.name" -> "name") and true if the entry already
// exists, false if it's not registered.
func walk(currentCategory object, key string) (object, string, bool) { // TODO test walk more extensively
	b := 0
	e := strings.Index(key, ".")
	if e == -1 {
		e = len(key)
	}
	for catKey := key[b:e]; ; catKey = key[b:e] {
		entry, ok := currentCategory[catKey]
		if !ok {
			// If categories are missing, create them
			currentCategory = createMissingCategories(currentCategory, key[b:])
			i := strings.LastIndex(key, ".")
			if i == -1 {
				catKey = key
			} else {
				catKey = key[i+1:]
			}

			// Entry doesn't exist and is not registered
			return currentCategory, catKey, false
		}

		if category, ok := entry.(object); ok {
			currentCategory = category
		} else {
			if e < len(key) {
				panic(fmt.Sprintf("Attempted to add an entry to non-category %q", key[:e]))
			}

			// Entry exists
			return currentCategory, catKey, true
		}

		if e+1 <= len(key) {
			b = e + 1
			newE := strings.Index(key[b:], ".")
			if newE == -1 {
				e = len(key)
			} else {
				e = newE + b
			}
		} else {
			break
		}
	}

	panic(fmt.Sprintf("Attempted to replace the %q category with an entry", key))
}

// createMissingCategories based on the key path, starting at the given index.
// Doesn't create anything is not needed.
// Returns the deepest category created, or the provided object if nothing has
// been created.
func createMissingCategories(currentCategory object, path string) object { // TODO create missing categories
	b := 0
	e := strings.Index(path, ".")
	if e == -1 {
		return currentCategory
	}
	for catKey := path[b:e]; ; catKey = path[b:e] {
		newCategory := object{}
		currentCategory[catKey] = newCategory
		currentCategory = newCategory

		if e+1 <= len(path) {
			b = e + 1
			newE := strings.Index(path[b:], ".")
			if newE == -1 {
				return currentCategory
			}
			e = newE + b
		} else {
			break
		}
	}
	return currentCategory
}

func loadDefaults(src object, dst object) {
	for k, v := range src {
		if obj, ok := v.(object); ok {
			sub := make(object, len(obj))
			loadDefaults(obj, sub)
			dst[k] = sub
		} else {
			entry := v.(*Entry)
			dst[k] = &Entry{entry.Value, entry.Type, entry.AuthorizedValues}
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
			// TODO check conflicts ?
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
