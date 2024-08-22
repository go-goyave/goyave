package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"strings"
	"sync"

	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

type object map[string]any

type readFunc func(string) (object, error)

// Config structure holding a configuration that should be used for a single
// instance of `goyave.Server`.
//
// This structure is not protected for safe concurrent access in order to increase
// performance. Therefore, you should never use the `Set()` function when the configuration
// is in use by an already running server.
type Config struct {
	config object
}

// Error returned when the configuration could not
// be loaded or is invalid.
// Can be unwraped to get the original error.
type Error struct {
	err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("Config error: %s", e.err.Error())
}

func (e *Error) Unwrap() error {
	return e.err
}

type loader struct {
	defaults object
	mu       sync.RWMutex
}

var defaultLoader = &loader{
	defaults: configDefaults,
}

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
	defaultLoader.register(key, entry)
}

func (l *loader) register(key string, entry Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	category, entryKey, exists := walk(l.defaults, key)
	if exists {
		if !reflect.DeepEqual(&entry, category[entryKey].(*Entry)) {
			panic(errors.Errorf("attempted to override registered config entry %q", key))
		}
	} else {
		category[entryKey] = &entry
	}
}

func (l *loader) loadFrom(fs fs.FS, path string) (*Config, error) {
	return l.load(func(_ string) (object, error) { return l.readConfigFile(fs, path) }, path)
}

func (l *loader) loadJSON(cfg string) (*Config, error) {
	return l.load(l.readString, cfg)
}

func (l *loader) load(readFunc readFunc, source string) (*Config, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	config := make(object, len(l.defaults))
	loadDefaults(l.defaults, config)

	if readFunc != nil {
		conf, err := readFunc(source)
		if err != nil {
			return nil, errors.New(&Error{err})
		}

		if err := override(conf, config); err != nil {
			return nil, errors.New(&Error{err})
		}
	}

	if err := config.validate(""); err != nil {
		return nil, errors.New(&Error{err})
	}

	return &Config{
		config: config,
	}, nil
}

// Load loads the config.json file in the current working directory.
// If the "GOYAVE_ENV" env variable is set, the config file will be picked like so:
//   - "production": "config.production.json"
//   - "test": "config.test.json"
//   - By default: "config.json"
func Load() (*Config, error) {
	return defaultLoader.loadFrom(&osfs.FS{}, getConfigFilePath())
}

// LoadDefault loads default config.
func LoadDefault() *Config {
	cfg, _ := defaultLoader.load(nil, "")
	return cfg
}

// LoadFrom loads a config file from the given path.
func LoadFrom(path string) (*Config, error) {
	return defaultLoader.loadFrom(&osfs.FS{}, path)
}

// LoadJSON load a configuration file from raw JSON. Can be used in combination with
// Go's embed directive.
//
//	var (
//		//go:embed config.json
//		cfgJSON string
//	)
//
//	func main() {
//		cfg, err := config.LoadJSON(cfgJSON)
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err.(*errors.Error).String())
//			os.Exit(1)
//		}
//
//		server, err := goyave.New(goyave.Options{Config: cfg})
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err.(*errors.Error).String())
//			os.Exit(1)
//		}
//
//		// ...
//	}
func LoadJSON(cfg string) (*Config, error) {
	return defaultLoader.loadJSON(cfg)
}

func getConfigFilePath() string {
	env := strings.ToLower(os.Getenv("GOYAVE_ENV"))
	if env == "local" || env == "localhost" || env == "" {
		return "config.json"
	}
	return "config." + env + ".json"
}

func (l *loader) readConfigFile(filesystem fs.FS, file string) (o object, err error) {
	var configFile fs.File
	o = make(object, len(l.defaults))
	configFile, err = filesystem.Open(file)

	if err == nil {
		defer func() {
			e := configFile.Close()
			if err == nil && e != nil {
				err = errors.New(e)
			}
		}()
		jsonParser := json.NewDecoder(configFile)
		err = errors.New(jsonParser.Decode(&o))
	} else {
		err = errors.New(err)
	}

	return
}

func (l *loader) readString(str string) (object, error) {
	conf := make(object, len(l.defaults))
	if err := json.NewDecoder(strings.NewReader(str)).Decode(&conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// walk the config using the key. Returns the deepest category, the entry key
// with its path stripped ("app.name" -> "name") and true if the entry already
// exists, false if it's not registered.
func walk(currentCategory object, key string) (object, string, bool) {
	if key == "" {
		panic(errors.New("empty key is not allowed"))
	}

	if key[len(key)-1:] == "." {
		panic(errors.New("keys ending with a dot are not allowed"))
	}

	start := 0
	dotIndex := strings.Index(key, ".")
	if dotIndex == -1 {
		dotIndex = len(key)
	}
	for catKey := key[start:dotIndex]; ; catKey = key[start:dotIndex] {
		entry, ok := currentCategory[catKey]
		if !ok {
			// If categories are missing, create them
			currentCategory = createMissingCategories(currentCategory, key[start:])
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
			if dotIndex < len(key) {
				panic(errors.Errorf("attempted to add an entry to non-category %q", key[:dotIndex]))
			}

			// Entry exists
			return currentCategory, catKey, true
		}

		if dotIndex+1 <= len(key) {
			start = dotIndex + 1
			newDotIndex := strings.Index(key[start:], ".")
			if newDotIndex == -1 {
				dotIndex = len(key)
			} else {
				dotIndex = newDotIndex + start
			}
		} else {
			break
		}
	}

	panic(errors.Errorf("attempted to replace the %q category with an entry", key))
}

// createMissingCategories based on the key path, starting at the given index.
// Doesn't create anything is not needed.
// Returns the deepest category created, or the provided object if nothing has
// been created.
func createMissingCategories(currentCategory object, path string) object {
	start := 0
	dotIndex := strings.Index(path, ".")
	if dotIndex == -1 {
		return currentCategory
	}
	for catKey := path[start:dotIndex]; ; catKey = path[start:dotIndex] {
		newCategory := object{}
		currentCategory[catKey] = newCategory
		currentCategory = newCategory

		if dotIndex+1 <= len(path) {
			start = dotIndex + 1
			newDotIndex := strings.Index(path[start:], ".")
			if newDotIndex == -1 {
				return currentCategory
			}
			dotIndex = newDotIndex + start
		}
	}
}

func override(src object, dst object) error {
	for k, v := range src {
		if obj, ok := v.(map[string]any); ok {
			if dstObj, ok := dst[k]; !ok {
				dst[k] = make(object, len(obj))
			} else if _, ok := dstObj.(object); !ok {
				// Conflict: destination is not a category
				return fmt.Errorf("\n\t- cannot override entry %q with a category", k)
			}
			if err := override(obj, dst[k].(object)); err != nil {
				return err
			}
		} else if entry, ok := dst[k]; ok {
			e, ok := entry.(*Entry)
			if !ok {
				// Conflict: override category with an entry
				return fmt.Errorf("\n\t- cannot override category %q with an entry", k)
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
		return fmt.Errorf("%s", message)
	}
	return nil
}

// Get a config entry using a dot-separated path.
// Panics if the entry doesn't exist.
func (c *Config) Get(key string) any {
	if val, ok := c.get(key); ok {
		return val
	}

	panic(errors.Errorf("config entry \"%s\" doesn't exist", key))
}

func (c *Config) get(key string) (any, bool) {
	currentCategory := c.config
	start := 0
	dotIndex := strings.Index(key, ".")
	if dotIndex == -1 {
		dotIndex = len(key)
	}
	for path := key[start:dotIndex]; ; path = key[start:dotIndex] {
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

		if dotIndex+1 <= len(key) {
			start = dotIndex + 1
			newDotIndex := strings.Index(key[start:], ".")
			if newDotIndex == -1 {
				dotIndex = len(key)
			} else {
				dotIndex = newDotIndex + start
			}
		}
	}
	return nil, false
}

// GetString a config entry as string.
// Panics if entry is not a string or if it doesn't exist.
func (c *Config) GetString(key string) string {
	str, ok := c.Get(key).(string)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a string", key))
	}
	return str
}

// GetBool a config entry as bool.
// Panics if entry is not a bool or if it doesn't exist.
func (c *Config) GetBool(key string) bool {
	val, ok := c.Get(key).(bool)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a bool", key))
	}
	return val
}

// GetInt a config entry as int.
// Panics if entry is not an int or if it doesn't exist.
func (c *Config) GetInt(key string) int {
	val, ok := c.Get(key).(int)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not an int", key))
	}
	return val
}

// GetFloat a config entry as float64.
// Panics if entry is not a float64 or if it doesn't exist.
func (c *Config) GetFloat(key string) float64 {
	val, ok := c.Get(key).(float64)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a float64", key))
	}
	return val
}

// GetStringSlice a config entry as []string.
// Panics if entry is not a string slice or if it doesn't exist.
func (c *Config) GetStringSlice(key string) []string {
	str, ok := c.Get(key).([]string)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a string slice", key))
	}
	return str
}

// GetBoolSlice a config entry as []bool.
// Panics if entry is not a bool slice or if it doesn't exist.
func (c *Config) GetBoolSlice(key string) []bool {
	str, ok := c.Get(key).([]bool)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a bool slice", key))
	}
	return str
}

// GetIntSlice a config entry as []int.
// Panics if entry is not an int slice or if it doesn't exist.
func (c *Config) GetIntSlice(key string) []int {
	str, ok := c.Get(key).([]int)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not an int slice", key))
	}
	return str
}

// GetFloatSlice a config entry as []float64.
// Panics if entry is not a float slice or if it doesn't exist.
func (c *Config) GetFloatSlice(key string) []float64 {
	str, ok := c.Get(key).([]float64)
	if !ok {
		panic(errors.Errorf("config entry \"%s\" is not a float64 slice", key))
	}
	return str
}

// Has check if a config entry exists.
func (c *Config) Has(key string) bool {
	_, ok := c.get(key)
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
//
// This operation is not concurrently safe and should not be used when the configuration
// is in use by an already running server.
func (c *Config) Set(key string, value any) {
	category, entryKey, exists := walk(c.config, key)
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
