package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"goyave.dev/goyave/v4/util/sliceutil"
)

type object map[string]interface{}

// Entry is the internal reprensentation of a config entry.
// It contains the entry value, its expected type (for validation)
// and a slice of authorized values (for validation too). If this slice
// is empty, it means any value can be used, provided it is of the correct type.
type Entry struct {
	Value            interface{}
	AuthorizedValues []interface{} // Leave empty for "any"
	Type             reflect.Kind
	IsSlice          bool
}

type readFunc func(string) (object, error)

var config object

var configDefaults = object{
	"app": object{
		"name":            &Entry{"goyave", []interface{}{}, reflect.String, false},
		"environment":     &Entry{"localhost", []interface{}{}, reflect.String, false},
		"debug":           &Entry{true, []interface{}{}, reflect.Bool, false},
		"defaultLanguage": &Entry{"en-US", []interface{}{}, reflect.String, false},
	},
	"server": object{
		"host":          &Entry{"127.0.0.1", []interface{}{}, reflect.String, false},
		"domain":        &Entry{"", []interface{}{}, reflect.String, false},
		"protocol":      &Entry{"http", []interface{}{"http", "https"}, reflect.String, false},
		"port":          &Entry{8080, []interface{}{}, reflect.Int, false},
		"httpsPort":     &Entry{8081, []interface{}{}, reflect.Int, false},
		"timeout":       &Entry{10, []interface{}{}, reflect.Int, false},
		"maxUploadSize": &Entry{10.0, []interface{}{}, reflect.Float64, false},
		"maintenance":   &Entry{false, []interface{}{}, reflect.Bool, false},
		"tls": object{
			"cert": &Entry{nil, []interface{}{}, reflect.String, false},
			"key":  &Entry{nil, []interface{}{}, reflect.String, false},
		},
		"proxy": object{
			"protocol": &Entry{"http", []interface{}{"http", "https"}, reflect.String, false},
			"host":     &Entry{nil, []interface{}{}, reflect.String, false},
			"port":     &Entry{80, []interface{}{}, reflect.Int, false},
			"base":     &Entry{"", []interface{}{}, reflect.String, false},
		},
	},
	"database": object{
		"connection":         &Entry{"none", []interface{}{}, reflect.String, false},
		"host":               &Entry{"127.0.0.1", []interface{}{}, reflect.String, false},
		"port":               &Entry{3306, []interface{}{}, reflect.Int, false},
		"name":               &Entry{"goyave", []interface{}{}, reflect.String, false},
		"username":           &Entry{"root", []interface{}{}, reflect.String, false},
		"password":           &Entry{"root", []interface{}{}, reflect.String, false},
		"options":            &Entry{"charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", []interface{}{}, reflect.String, false},
		"maxOpenConnections": &Entry{20, []interface{}{}, reflect.Int, false},
		"maxIdleConnections": &Entry{20, []interface{}{}, reflect.Int, false},
		"maxLifetime":        &Entry{300, []interface{}{}, reflect.Int, false},
		"autoMigrate":        &Entry{false, []interface{}{}, reflect.Bool, false},
		"config": object{
			"skipDefaultTransaction":                   &Entry{false, []interface{}{}, reflect.Bool, false},
			"dryRun":                                   &Entry{false, []interface{}{}, reflect.Bool, false},
			"prepareStmt":                              &Entry{true, []interface{}{}, reflect.Bool, false},
			"disableNestedTransaction":                 &Entry{false, []interface{}{}, reflect.Bool, false},
			"allowGlobalUpdate":                        &Entry{false, []interface{}{}, reflect.Bool, false},
			"disableAutomaticPing":                     &Entry{false, []interface{}{}, reflect.Bool, false},
			"disableForeignKeyConstraintWhenMigrating": &Entry{false, []interface{}{}, reflect.Bool, false},
		},
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
func Register(key string, entry Entry) {
	mutex.Lock()
	defer mutex.Unlock()
	category, entryKey, exists := walk(configDefaults, key)
	if exists {
		if !reflect.DeepEqual(&entry, category[entryKey].(*Entry)) {
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
func Load() error {
	return LoadFrom(getConfigFilePath())
}

// LoadFrom loads a config file from the given path.
func LoadFrom(path string) error {
	return load(readConfigFile, path)
}

// LoadJSON load a configuration file from raw JSON. Can be used in combination with
// Go's 1.16 embed directive.
//
//	 var (
//	 	//go:embed config.json
//	 	cfg string
//	 )
//
//	 func main() {
//	 	if err := config.LoadJSON(cfg); err != nil {
//	 		goyave.ErrLogger.Println(err)
//	 		os.Exit(goyave.ExitInvalidConfig)
//	 	}
//
//	 	if err := goyave.Start(route.Register); err != nil {
//	 		os.Exit(err.(*goyave.Error).ExitCode)
//		 }
//	 }
func LoadJSON(cfg string) error {
	return load(readString, cfg)
}

func load(readFunc readFunc, source string) error {
	mutex.Lock()
	defer mutex.Unlock()
	config = make(object, len(configDefaults))
	loadDefaults(configDefaults, config)

	conf, err := readFunc(source)
	if err != nil {
		config = nil
		return err
	}

	if err := override(conf, config); err != nil {
		config = nil
		return err
	}

	if err := config.validate(""); err != nil {
		config = nil
		return fmt.Errorf("Invalid config:%s", err.Error())
	}

	return nil
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

// Get a config entry. Panics if the entry doesn't exist.
func Get(key string) interface{} {
	if val, ok := get(key); ok {
		return val
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

func get(key string) (interface{}, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	if config == nil {
		panic("Config is not loaded")
	}
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

// GetString a config entry as string.
// Panics if entry is not a string or if it doesn't exist.
func GetString(key string) string {
	str, ok := Get(key).(string)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a string", key))
	}
	return str
}

// GetBool a config entry as bool.
// Panics if entry is not a bool or if it doesn't exist.
func GetBool(key string) bool {
	val, ok := Get(key).(bool)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a bool", key))
	}
	return val
}

// GetInt a config entry as int.
// Panics if entry is not an int or if it doesn't exist.
func GetInt(key string) int {
	val, ok := Get(key).(int)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not an int", key))
	}
	return val
}

// GetFloat a config entry as float64.
// Panics if entry is not a float64 or if it doesn't exist.
func GetFloat(key string) float64 {
	val, ok := Get(key).(float64)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a float64", key))
	}
	return val
}

// GetStringSlice a config entry as []string.
// Panics if entry is not a string slice or if it doesn't exist.
func GetStringSlice(key string) []string {
	str, ok := Get(key).([]string)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a string slice", key))
	}
	return str
}

// GetBoolSlice a config entry as []bool.
// Panics if entry is not a bool slice or if it doesn't exist.
func GetBoolSlice(key string) []bool {
	str, ok := Get(key).([]bool)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a bool slice", key))
	}
	return str
}

// GetIntSlice a config entry as []int.
// Panics if entry is not an int slice or if it doesn't exist.
func GetIntSlice(key string) []int {
	str, ok := Get(key).([]int)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not an int slice", key))
	}
	return str
}

// GetFloatSlice a config entry as []float64.
// Panics if entry is not a float slice or if it doesn't exist.
func GetFloatSlice(key string) []float64 {
	str, ok := Get(key).([]float64)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a float64 slice", key))
	}
	return str
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
//   - A category cannot be replaced with an entry.
//   - An entry cannot be replaced with a category.
//   - New categories can be created with they don't already exist.
//   - New entries can be created if they don't already exist. This new entry
//     will be subsequently validated using the type of its initial value and
//     have an empty slice as authorized values (meaning it can have any value of its type)
//
// Panics and revert changes in case of error.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	if config == nil {
		panic("Config is not loaded")
	}
	category, entryKey, exists := walk(config, key)
	if exists {
		entry := category[entryKey].(*Entry)
		previous := entry.Value
		entry.Value = value
		if err := entry.validate(key); err != nil {
			entry.Value = previous
			panic(err)
		}
		category[entryKey] = entry
	} else {
		category[entryKey] = makeEntryFromValue(value)
	}
}

// walk the config using the key. Returns the deepest category, the entry key
// with its path stripped ("app.name" -> "name") and true if the entry already
// exists, false if it's not registered.
func walk(currentCategory object, key string) (object, string, bool) {
	if key == "" {
		panic("Empty key is not allowed")
	}

	if key[len(key)-1:] == "." {
		panic("Keys ending with a dot are not allowed")
	}

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
func createMissingCategories(currentCategory object, path string) object {
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
			entry := v.(*Entry)
			value := entry.Value
			t := reflect.TypeOf(value)
			if t != nil && t.Kind() == reflect.Slice {
				list := reflect.ValueOf(value)
				length := list.Len()
				slice := reflect.MakeSlice(reflect.SliceOf(t.Elem()), 0, length)
				for i := 0; i < length; i++ {
					slice = reflect.Append(slice, list.Index(i))
				}
				value = slice.Interface()
			}
			dst[k] = &Entry{value, entry.AuthorizedValues, entry.Type, entry.IsSlice}
		}
	}
}

func override(src object, dst object) error {
	for k, v := range src {
		if obj, ok := v.(map[string]interface{}); ok {
			if dstObj, ok := dst[k]; !ok {
				dst[k] = make(object, len(obj))
			} else if _, ok := dstObj.(object); !ok {
				// Conflict: destination is not a category
				return fmt.Errorf("Invalid config:\n\t- Cannot override entry %q with a category", k)
			}
			if err := override(obj, dst[k].(object)); err != nil {
				return err
			}
		} else if entry, ok := dst[k]; ok {
			e, ok := entry.(*Entry)
			if !ok {
				// Conflict: override category with an entry
				return fmt.Errorf("Invalid config:\n\t- Cannot override category %q with an entry", k)
			}
			e.Value = v
		} else {
			// If entry doesn't exist (and is not registered),
			// register it with the type of the type given here
			// and "any" authorized values.
			dst[k] = makeEntryFromValue(v)
		}
	}
	return nil
}

func makeEntryFromValue(value interface{}) *Entry {
	isSlice := false
	t := reflect.TypeOf(value)
	kind := t.Kind()
	if kind == reflect.Slice {
		kind = t.Elem().Kind()
		isSlice = true
	}
	return &Entry{value, []interface{}{}, kind, isSlice}
}

func readConfigFile(file string) (object, error) {
	conf := make(object, len(configDefaults))
	configFile, err := os.Open(file)

	if err == nil {
		defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		err = jsonParser.Decode(&conf)
	}
	return conf, err
}

func readString(str string) (object, error) {
	conf := make(object, len(configDefaults))
	if err := json.NewDecoder(strings.NewReader(str)).Decode(&conf); err != nil {
		return nil, err
	}
	return conf, nil
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

	if err := e.tryEnvVarConversion(key); err != nil {
		return err
	}

	t := reflect.TypeOf(e.Value)
	kind := t.Kind()
	if e.IsSlice && kind == reflect.Slice {
		kind = t.Elem().Kind()
	}
	if kind != e.Type {
		if !e.tryIntConversion(kind) {
			var message string
			if e.IsSlice {
				message = "%q must be a slice of %s"
			} else {
				message = "%q type must be %s"
			}

			return fmt.Errorf(message, key, e.Type)
		}
		return nil
	}

	if len(e.AuthorizedValues) > 0 {
		if e.IsSlice {
			// Accepted values for slices define the values that can be used inside the slice
			// It doesn't represent the value of the slice itself (content and order)
			list := reflect.ValueOf(e.Value)
			length := list.Len()
			authorizedValuesList := reflect.ValueOf(e.AuthorizedValues)
			for i := 0; i < length; i++ {
				if !e.authorizedValuesContains(authorizedValuesList, list.Index(i).Interface()) {
					return fmt.Errorf("%q elements must have one of the following values: %v", key, e.AuthorizedValues)
				}
			}
		} else if !sliceutil.Contains(e.AuthorizedValues, e.Value) {
			return fmt.Errorf("%q must have one of the following values: %v", key, e.AuthorizedValues)
		}
	}

	return nil
}

// authorizedValuesContains avoids to recreate the reflect.Value of the list for every check
func (e *Entry) authorizedValuesContains(list reflect.Value, value interface{}) bool {
	length := list.Len()
	for i := 0; i < length; i++ {
		if list.Index(i).Interface() == value {
			return true
		}
	}
	return false
}

func (e *Entry) tryIntConversion(kind reflect.Kind) bool {
	if kind == reflect.Float64 && e.Type == reflect.Int {
		if e.IsSlice {
			return e.convertIntSlice()
		}

		intVal, ok := e.convertInt(e.Value.(float64))
		if ok {
			e.Value = intVal
			return true
		}
	}

	return false
}

func (e *Entry) convertInt(value float64) (int, bool) {
	intVal := int(value)
	if value == float64(intVal) {
		return intVal, true
	}
	return 0, false
}

func (e *Entry) convertIntSlice() bool {
	original := e.Value.([]float64)
	slice := make([]int, len(original))
	for k, v := range original {
		intVal, ok := e.convertInt(v)
		if !ok {
			return false
		}
		slice[k] = intVal
	}
	e.Value = slice
	return true
}

func (e *Entry) tryEnvVarConversion(key string) error {
	str, ok := e.Value.(string)
	if ok {
		val, err := e.convertEnvVar(str, key)
		if err == nil && val != nil {
			e.Value = val
		}
		return err
	}

	return nil
}

func (e *Entry) convertEnvVar(str, key string) (interface{}, error) {
	if strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
		varName := str[2 : len(str)-1]
		value, set := os.LookupEnv(varName)
		if !set {
			return nil, fmt.Errorf("%q: %q environment variable is not set", key, varName)
		}

		switch e.Type {
		case reflect.Int:
			if i, err := strconv.Atoi(value); err == nil {
				return i, nil
			}
			return nil, fmt.Errorf("%q could not be converted to int from environment variable %q of value %q", key, varName, value)
		case reflect.Float64:
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				return f, nil
			}
			return nil, fmt.Errorf("%q could not be converted to float64 from environment variable %q of value %q", key, varName, value)
		case reflect.Bool:
			if b, err := strconv.ParseBool(value); err == nil {
				return b, nil
			}
			return nil, fmt.Errorf("%q could not be converted to bool from environment variable %q of value %q", key, varName, value)
		default:
			// Keep value as string if type is not supported and let validation do its job
			return value, nil
		}
	}

	return nil, nil
}
