package config

import (
	"fmt"
	"strings"
)

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

func LoadV5() (*Config, error) {
	return LoadFromV5(getConfigFilePath())
}

func LoadFromV5(path string) (*Config, error) {
	return loadV5(readConfigFile, path)
}

func LoadJSONV5(cfg string) error {
	return load(readString, cfg)
}

func loadV5(readFunc readFunc, source string) (*Config, error) {
	config := make(object, len(configDefaults))
	loadDefaults(configDefaults, config)

	conf, err := readFunc(source)
	if err != nil {
		return nil, &Error{err}
	}

	if err := override(conf, config); err != nil {
		return nil, &Error{err}
	}

	if err := config.validate(""); err != nil {
		return nil, &Error{err}
	}

	return &Config{
		config: config,
	}, nil
}

// Get a config entry. Panics if the entry doesn't exist.
func (c *Config) Get(key string) any {
	if val, ok := c.get(key); ok {
		return val
	}

	panic(fmt.Sprintf("Config entry \"%s\" doesn't exist", key))
}

func (c *Config) get(key string) (any, bool) {
	currentCategory := c.config
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
func (c *Config) GetString(key string) string {
	str, ok := c.Get(key).(string)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a string", key))
	}
	return str
}

// GetBool a config entry as bool.
// Panics if entry is not a bool or if it doesn't exist.
func (c *Config) GetBool(key string) bool {
	val, ok := c.Get(key).(bool)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a bool", key))
	}
	return val
}

// GetInt a config entry as int.
// Panics if entry is not an int or if it doesn't exist.
func (c *Config) GetInt(key string) int {
	val, ok := c.Get(key).(int)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not an int", key))
	}
	return val
}

// GetFloat a config entry as float64.
// Panics if entry is not a float64 or if it doesn't exist.
func (c *Config) GetFloat(key string) float64 {
	val, ok := c.Get(key).(float64)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a float64", key))
	}
	return val
}

// GetStringSlice a config entry as []string.
// Panics if entry is not a string slice or if it doesn't exist.
func (c *Config) GetStringSlice(key string) []string {
	str, ok := c.Get(key).([]string)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a string slice", key))
	}
	return str
}

// GetBoolSlice a config entry as []bool.
// Panics if entry is not a bool slice or if it doesn't exist.
func (c *Config) GetBoolSlice(key string) []bool {
	str, ok := c.Get(key).([]bool)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a bool slice", key))
	}
	return str
}

// GetIntSlice a config entry as []int.
// Panics if entry is not an int slice or if it doesn't exist.
func (c *Config) GetIntSlice(key string) []int {
	str, ok := c.Get(key).([]int)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not an int slice", key))
	}
	return str
}

// GetFloatSlice a config entry as []float64.
// Panics if entry is not a float slice or if it doesn't exist.
func (c *Config) GetFloatSlice(key string) []float64 {
	str, ok := c.Get(key).([]float64)
	if !ok {
		panic(fmt.Sprintf("Config entry \"%s\" is not a float64 slice", key))
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
//  - A category cannot be replaced with an entry.
//  - An entry cannot be replaced with a category.
//  - New categories can be created with they don't already exist.
//  - New entries can be created if they don't already exist. This new entry
//    will be subsequently validated using the type of its initial value and
//    have an empty slice as authorized values (meaning it can have any value of its type)
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
