package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigError(t *testing.T) {
	e := &Error{
		err: fmt.Errorf("test error"),
	}

	assert.Equal(t, "Config error: test error", e.Error())
	assert.Equal(t, e.err, e.Unwrap())
}

func TestGetConfigFilePath(t *testing.T) {
	t.Setenv("GOYAVE_ENV", "localhost")
	assert.Equal(t, "config.json", getConfigFilePath())

	t.Setenv("GOYAVE_ENV", "test")
	assert.Equal(t, "config.test.json", getConfigFilePath())

	t.Setenv("GOYAVE_ENV", "production")
	assert.Equal(t, "config.production.json", getConfigFilePath())
}

func TestRegister(t *testing.T) {
	entry := Entry{
		Value:            "",
		AuthorizedValues: []any{},
		Type:             reflect.String,
		IsSlice:          false,
	}
	Register("testEntry", entry)
	assert.Equal(t, &entry, defaultLoader.defaults["testEntry"])
}

func TestLoad(t *testing.T) {
	t.Run("Load", func(t *testing.T) {
		// Should use automatically generated path (based on GOYAVE_ENV)
		t.Setenv("GOYAVE_ENV", "test")
		cfg, err := Load()
		require.NoError(t, err)

		assert.NotNil(t, cfg)

		expected := &Entry{
			Value:            "root level content",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		}
		assert.Equal(t, expected, cfg.config["rootLevel"])

		// Default config also loaded
		expected = &Entry{
			Value:            "goyave",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
			Required:         true,
		}
		assert.Equal(t, expected, cfg.config["app"].(object)["name"])
	})

	t.Run("Load Invalid", func(t *testing.T) {
		t.Setenv("GOYAVE_ENV", "test_invalid")
		cfg, err := Load()
		assert.Nil(t, cfg)
		require.Error(t, err)
	})

	t.Run("Load Default", func(t *testing.T) {
		cfg := LoadDefault()
		assert.Equal(t, defaultLoader.defaults, cfg.config)
	})

	t.Run("Load Non Existing", func(t *testing.T) {
		t.Setenv("GOYAVE_ENV", "nonexisting")
		cfg, err := Load()
		assert.Nil(t, cfg)
		require.Error(t, err)
	})

	t.Run("LoadFrom", func(t *testing.T) {
		cfg, err := LoadFrom("../resources/custom_config.json")
		require.NoError(t, err)

		assert.NotNil(t, cfg)

		expected := &Entry{
			Value:            "value",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		}
		assert.Equal(t, expected, cfg.config["custom-entry"])

		// Default config also loaded
		expected = &Entry{
			Value:            "goyave",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
			Required:         true,
		}
		assert.Equal(t, expected, cfg.config["app"].(object)["name"])
	})

	t.Run("LoadJSON", func(t *testing.T) {
		cfg, err := LoadJSON(`{"custom-entry": "value"}`)
		require.NoError(t, err)

		assert.NotNil(t, cfg)

		expected := &Entry{
			Value:            "value",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		}
		assert.Equal(t, expected, cfg.config["custom-entry"])

		// Default config also loaded
		expected = &Entry{
			Value:            "goyave",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
			Required:         true,
		}
		assert.Equal(t, expected, cfg.config["app"].(object)["name"])
	})

	t.Run("LoadJSON Invalid", func(t *testing.T) {
		cfg, err := LoadJSON(`{"unclosed":`)
		assert.Nil(t, cfg)
		require.Error(t, err)
	})

	t.Run("Load Override Entry With Category", func(t *testing.T) {
		cfg, err := LoadJSON(`{"app": {"name": {}}}`)
		assert.Nil(t, cfg)
		require.Error(t, err)
		assert.Equal(t, "Config error: \n\t- cannot override entry \"name\" with a category", err.Error())
	})

	t.Run("Load Override Category With Entry", func(t *testing.T) {
		cfg, err := LoadJSON(`{"app": "value"}`)
		assert.Nil(t, cfg)
		require.Error(t, err)
		assert.Equal(t, "Config error: \n\t- cannot override category \"app\" with an entry", err.Error())
	})

	t.Run("Validation", func(t *testing.T) {
		cfg, err := LoadJSON(`{"app": {"name": 123}}`)
		assert.Nil(t, cfg)
		require.Error(t, err)
		assert.Equal(t, "Config error: \n\t- \"app.name\" type must be string", err.Error())
	})

	t.Run("Load Env Variables", func(t *testing.T) {
		defaultLoader.mu.Lock()
		loader := loader{
			defaults: make(object, len(defaultLoader.defaults)),
		}
		loadDefaults(defaultLoader.defaults, loader.defaults)
		defaultLoader.mu.Unlock()

		loader.register("envString", Entry{
			Value:            "",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		})
		loader.register("envInt", Entry{
			Value:            0,
			AuthorizedValues: []any{},
			Type:             reflect.Int,
			IsSlice:          false,
		})
		loader.register("envFloat", Entry{
			Value:            0.0,
			AuthorizedValues: []any{},
			Type:             reflect.Float64,
			IsSlice:          false,
		})
		loader.register("envBool", Entry{
			Value:            false,
			AuthorizedValues: []any{},
			Type:             reflect.Bool,
			IsSlice:          false,
		})

		t.Setenv("TEST_ENV_STRING", "hello")
		t.Setenv("TEST_ENV_INT", "123")
		t.Setenv("TEST_ENV_FLOAT", "123.456")
		t.Setenv("TEST_ENV_BOOL", "TRUE")

		json := `{
			"envString": "${TEST_ENV_STRING}",
			"envInt": "${TEST_ENV_INT}",
			"envFloat": "${TEST_ENV_FLOAT}",
			"envBool": "${TEST_ENV_BOOL}"
		}`

		cfg, err := loader.loadJSON(json)
		require.NoError(t, err)

		assert.Equal(t, "hello", cfg.Get("envString"))
		assert.Equal(t, 123, cfg.Get("envInt"))
		assert.InEpsilon(t, 123.456, cfg.Get("envFloat"), 0)
		assert.Equal(t, true, cfg.Get("envBool"))

		// Invalid int
		t.Setenv("TEST_ENV_INT", "hello")
		cfg, err = loader.loadJSON(json)
		require.Error(t, err)
		assert.Nil(t, cfg)
		t.Setenv("TEST_ENV_INT", "123")

		// Invalid float
		t.Setenv("TEST_ENV_FLOAT", "hello")
		cfg, err = loader.loadJSON(json)
		require.Error(t, err)
		assert.Nil(t, cfg)
		t.Setenv("TEST_ENV_FLOAT", "123.456")

		// Invalid bool
		t.Setenv("TEST_ENV_BOOL", "hello")
		cfg, err = loader.loadJSON(json)
		require.Error(t, err)
		assert.Nil(t, cfg)
		t.Setenv("TEST_ENV_BOOL", "TRUE")
	})

	t.Run("Load Env Variables Unsupported", func(t *testing.T) {
		loader := loader{
			defaults: make(object, len(defaultLoader.defaults)),
		}

		loader.register("envUnsupported", Entry{
			Value:            []string{},
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          true,
		})

		t.Setenv("TEST_ENV_UNSUPPORTED", "[hello]")

		json := `{
			"envUnsupported": "${TEST_ENV_UNSUPPORTED}"
		}`

		cfg, err := loader.loadJSON(json)
		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Load Env Variables Missing", func(t *testing.T) {
		defaultLoader.mu.Lock()
		loader := loader{
			defaults: make(object, len(defaultLoader.defaults)),
		}
		loadDefaults(defaultLoader.defaults, loader.defaults)
		defaultLoader.mu.Unlock()

		loader.register("envUnset", Entry{
			Value:            "",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		})

		json := `{
			"envUnsupported": "${TEST_ENV_UNSET}"
		}`

		cfg, err := loader.loadJSON(json)
		require.Error(t, err)
		assert.Nil(t, cfg)
	})

	t.Run("Create Missing Categories", func(t *testing.T) {
		json := `{
			"category": {
				"entry": 123,
				"subcategory": {
					"subentry": 456,
					"array": ["a", "b"],
					"deep": {
						"deepEntry": "deepValue"
					}
				}
			}
		}`
		cfg, err := LoadJSON(json)
		require.NoError(t, err)

		assert.NotNil(t, cfg)

		cat, ok := cfg.config["category"].(object)
		if !assert.True(t, ok) {
			return
		}
		expected := &Entry{
			Value:            123.0,
			AuthorizedValues: []any{},
			// It is float here because we haven't registered the config entry, so no conversion
			Type:    reflect.Float64,
			IsSlice: false,
		}
		assert.Equal(t, expected, cat["entry"])

		subcat, ok := cat["subcategory"].(object)
		if !assert.True(t, ok) {
			return
		}
		expected = &Entry{
			Value:            456.0,
			AuthorizedValues: []any{},
			Type:             reflect.Float64,
			IsSlice:          false,
		}
		assert.Equal(t, expected, subcat["subentry"])
		expected = &Entry{
			Value:            []any{"a", "b"},
			AuthorizedValues: []any{},
			Type:             reflect.Interface,
			IsSlice:          true,
		}
		assert.Equal(t, expected, subcat["array"])

		deep, ok := subcat["deep"].(object)
		if !assert.True(t, ok) {
			return
		}
		expected = &Entry{
			Value:            "deepValue",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		}
		assert.Equal(t, expected, deep["deepEntry"])

		// With partial existence
		cfg, err = LoadJSON(`{"app": {"subcategory": {"subentry": 456}}}`)
		require.NoError(t, err)

		assert.NotNil(t, cfg)

		cat, ok = cfg.config["app"].(object)
		if !assert.True(t, ok) {
			return
		}

		subcat, ok = cat["subcategory"].(object)
		if !assert.True(t, ok) {
			return
		}
		expected = &Entry{
			Value:            456.0,
			AuthorizedValues: []any{},
			Type:             reflect.Float64,
			IsSlice:          false,
		}
		assert.Equal(t, expected, subcat["subentry"])
	})
}

func TestConfig(t *testing.T) {
	defaultLoader.mu.Lock()
	loader := loader{
		defaults: make(object, len(defaultLoader.defaults)),
	}
	loadDefaults(defaultLoader.defaults, loader.defaults)
	defaultLoader.mu.Unlock()

	loader.register("testCategory.string", Entry{
		Value:            "",
		AuthorizedValues: []any{},
		Type:             reflect.String,
		IsSlice:          false,
	})
	loader.register("testCategory.int", Entry{
		Value:            0,
		AuthorizedValues: []any{},
		Type:             reflect.Int,
		IsSlice:          false,
	})
	loader.register("testCategory.float", Entry{
		Value:            0.0,
		AuthorizedValues: []any{},
		Type:             reflect.Float64,
		IsSlice:          false,
	})
	loader.register("testCategory.bool", Entry{
		Value:            false,
		AuthorizedValues: []any{},
		Type:             reflect.Bool,
		IsSlice:          false,
	})
	loader.register("testCategory.stringSlice", Entry{
		Value:            []string{},
		AuthorizedValues: []any{},
		Type:             reflect.String,
		IsSlice:          true,
	})
	loader.register("testCategory.intSlice", Entry{
		Value:            []int{},
		AuthorizedValues: []any{},
		Type:             reflect.Int,
		IsSlice:          true,
	})
	loader.register("testCategory.defaultIntSlice", Entry{
		Value:            []int{0, 1},
		AuthorizedValues: []any{},
		Type:             reflect.Int,
		IsSlice:          true,
	})
	loader.register("testCategory.floatSlice", Entry{
		Value:            []float64{},
		AuthorizedValues: []any{},
		Type:             reflect.Float64,
		IsSlice:          true,
	})
	loader.register("testCategory.boolSlice", Entry{
		Value:            []bool{},
		AuthorizedValues: []any{},
		Type:             reflect.Bool,
		IsSlice:          true,
	})

	loader.register("testCategory.set", Entry{
		Value:            0,
		AuthorizedValues: []any{456, 789},
		Type:             reflect.Int,
		IsSlice:          false,
	})

	loader.register("testCategory.setSlice", Entry{
		Value:            0,
		AuthorizedValues: []any{456, 789},
		Type:             reflect.Int,
		IsSlice:          true,
	})

	cfgJSON := `{
		"rootLevel": "root",
		"testCategory": {
			"string": "hello",
			"int": 123,
			"float": 123.456,
			"bool": true,
			"stringSlice": ["a", "b"],
			"intSlice": [1, 2],
			"floatSlice": [1.2, 3.4],
			"boolSlice": [true, false],
			"set": 456,
			"setSlice": []
		}
	}`

	cfg, err := loader.loadJSON(cfgJSON)
	require.NoError(t, err)

	t.Run("Get", func(t *testing.T) {
		v := cfg.Get("testCategory.int")
		assert.Equal(t, 123, v)

		v = cfg.Get("rootLevel")
		assert.Equal(t, "root", v)

		assert.Panics(t, func() {
			cfg.Get("testCategory.nonexistent")
		})
	})

	t.Run("Get Deep", func(t *testing.T) {
		cfg.Set("testCategory.subcategory.deep.entry", "hello")
		v := cfg.Get("testCategory.subcategory.deep.entry")
		assert.Equal(t, "hello", v)
	})

	t.Run("GetString", func(t *testing.T) {
		v := cfg.GetString("testCategory.string")
		assert.Equal(t, "hello", v)

		assert.Panics(t, func() {
			cfg.GetString("testCategory.int")
		})
	})

	t.Run("GetInt", func(t *testing.T) {
		v := cfg.GetInt("testCategory.int")
		assert.Equal(t, 123, v)

		assert.Panics(t, func() {
			cfg.GetInt("testCategory.string")
		})
	})

	t.Run("GetBool", func(t *testing.T) {
		v := cfg.GetBool("testCategory.bool")
		assert.True(t, v)

		assert.Panics(t, func() {
			cfg.GetBool("testCategory.string")
		})
	})

	t.Run("GetFloat", func(t *testing.T) {
		v := cfg.GetFloat("testCategory.float")
		assert.InEpsilon(t, 123.456, v, 0)

		assert.Panics(t, func() {
			cfg.GetFloat("testCategory.string")
		})
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		v := cfg.GetStringSlice("testCategory.stringSlice")
		assert.Equal(t, []string{"a", "b"}, v)

		assert.Panics(t, func() {
			cfg.GetStringSlice("testCategory.string")
		})
	})

	t.Run("GetBoolSlice", func(t *testing.T) {
		v := cfg.GetBoolSlice("testCategory.boolSlice")
		assert.Equal(t, []bool{true, false}, v)

		assert.Panics(t, func() {
			cfg.GetBoolSlice("testCategory.string")
		})
	})

	t.Run("GetIntSlice", func(t *testing.T) {
		v := cfg.GetIntSlice("testCategory.intSlice")
		assert.Equal(t, []int{1, 2}, v)

		assert.Panics(t, func() {
			cfg.GetIntSlice("testCategory.string")
		})

		v = cfg.GetIntSlice("testCategory.defaultIntSlice")
		assert.Equal(t, []int{0, 1}, v)
	})

	t.Run("GetFloatSlice", func(t *testing.T) {
		v := cfg.GetFloatSlice("testCategory.floatSlice")
		assert.Equal(t, []float64{1.2, 3.4}, v)

		assert.Panics(t, func() {
			cfg.GetFloatSlice("testCategory.string")
		})
	})

	t.Run("Has", func(t *testing.T) {
		assert.True(t, cfg.Has("testCategory.string"))
		assert.False(t, cfg.Has("testCategory.nonexistent"))
	})

	t.Run("Set", func(t *testing.T) {
		cfg.Set("testCategory.set", 789)
		expected := &Entry{
			Value:            789,
			AuthorizedValues: []any{456, 789},
			Type:             reflect.Int,
			IsSlice:          false,
		}
		assert.Equal(t, expected, cfg.config["testCategory"].(object)["set"])

		cfg.Set("testCategory.set", 456.0) // Conversion float->int
		expected = &Entry{
			Value:            456,
			AuthorizedValues: []any{456, 789},
			Type:             reflect.Int,
			IsSlice:          false,
		}
		assert.Equal(t, expected, cfg.config["testCategory"].(object)["set"])

		cfg.Set("testCategory.setSlice", []int{789, 456})
		expected = &Entry{
			Value:            []int{789, 456},
			AuthorizedValues: []any{456, 789},
			Type:             reflect.Int,
			IsSlice:          true,
		}
		assert.Equal(t, expected, cfg.config["testCategory"].(object)["setSlice"])

		// No need to validate the other conversions, they have been tested indirectly
		// through the loading at the start of this test

		assert.Panics(t, func() {
			cfg.Set("testCategory.intSlice", []any{1, "2"})
		})
		assert.Panics(t, func() {
			cfg.Set("testCategory.stringSlice", []any{1, "2"})
		})
		assert.Panics(t, func() {
			cfg.Set("testCategory.setSlice", "abc123")
		})
		assert.Panics(t, func() {
			cfg.Set("testCategory.set", "abc123")
		})
		assert.Panics(t, func() {
			// Unauthorized value
			cfg.Set("testCategory.set", 123)
		})

		assert.Panics(t, func() {
			// Unauthorized value
			cfg.Set("testCategory.setSlice", []int{456, 789, 123})
		})
	})

	t.Run("Set New Entry", func(t *testing.T) {
		cfg.Set("testCategory.subcategory.deep.entry", "hello")

		subcategory, ok := cfg.config["testCategory"].(object)["subcategory"].(object)["deep"].(object)
		if !assert.True(t, ok) {
			return
		}

		expected := &Entry{
			Value:            "hello",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
		}
		assert.Equal(t, expected, subcategory["entry"])
	})

	t.Run("Unset", func(t *testing.T) {
		cfg.Set("testCategory.set", nil)
		assert.Panics(t, func() {
			cfg.Get("testCategory.set")
		})
	})

	t.Run("Set Errors", func(t *testing.T) {
		assert.Panics(t, func() {
			cfg.Set("", "hello")
		})
		assert.Panics(t, func() {
			cfg.Set("testCategory.", "hello")
		})
	})

	t.Run("Set Override Entry With Category", func(t *testing.T) {
		assert.Panics(t, func() {
			cfg.Set("testCategory.string.entry", "hello")
		})
	})

	t.Run("Set Override Category With Entry", func(t *testing.T) {
		assert.Panics(t, func() {
			cfg.Set("testCategory", "hello")
		})
	})

	t.Run("Register Invalid", func(t *testing.T) {
		// Entry already exists and matches -> do nothing
		entry := Entry{
			Value:            "goyave",
			AuthorizedValues: []any{},
			Type:             reflect.String,
			IsSlice:          false,
			Required:         true,
		}
		loader.register("app.name", entry)
		assert.NotSame(t, &entry, loader.defaults["app"].(object)["name"])

		// Required values don't match (existing)
		assert.Panics(t, func() {
			loader.register("app.name", Entry{
				Value:            "goyave",
				AuthorizedValues: []any{"a", "b"},
				Type:             reflect.String,
				IsSlice:          true,
				Required:         true,
			})
		})

		// Type doesn't match (existing)
		assert.Panics(t, func() {
			loader.register("app.name", Entry{
				Value:            "goyave",
				AuthorizedValues: []any{},
				Type:             reflect.Int,
				IsSlice:          false,
				Required:         true,
			})
		})

		// Value doesn't match
		assert.Panics(t, func() {
			loader.register("app.name", Entry{
				Value:            "not goyave",
				AuthorizedValues: []any{},
				Type:             reflect.String,
				IsSlice:          false,
				Required:         true,
			})
		})
	})
}

func TestRequiredConfig(t *testing.T) {
	loader := loader{
		defaults: make(object),
	}

	loader.register("testCategory.nullValueNotRequired", Entry{
		Value:            nil,
		AuthorizedValues: []any{},
		Type:             reflect.String,
		Required:         false,
	})

	loader.register("testCategory.nullValueRequired", Entry{
		Value:            nil,
		AuthorizedValues: []any{},
		Type:             reflect.String,
		Required:         true,
	})

	loader.register("testCategory.valueCompletelyMissing", Entry{
		Type:     reflect.String,
		Required: true,
	})

	loader.register("testCategory.valueValidAndDefined", Entry{
		Type:     reflect.Int,
		Required: true,
	})

	var nilPointer *string
	loader.register("testCategory.nilPointerRequired", Entry{
		Value:    nilPointer,
		Type:     reflect.Ptr,
		Required: true,
	})

	validPointer := new(string)
	*validPointer = "valid"
	loader.register("testCategory.validPointer", Entry{
		Value:    validPointer,
		Type:     reflect.Ptr,
		Required: true,
	})

	cfgJSON := `{
		"testCategory": {
			"nullValueNotRequired": null,
			"nullValueRequired": null,
			"valueValidAndDefined": 123,
			"valueValidAndDefined": 123,
			"validPointer": "valid"
		}
	}`

	_, err := loader.loadJSON(cfgJSON)
	require.Error(t, err)

	expectedErrors := []string{
		"- \"testCategory.valueCompletelyMissing\" is required",
		"- \"testCategory.nullValueRequired\" is required",
		"- \"testCategory.nilPointerRequired\" is required",
	}

	actualErrors := strings.Split(err.Error(), "\n")
	for i, line := range actualErrors {
		actualErrors[i] = strings.TrimSpace(line)
	}

	for _, expectedMessage := range expectedErrors {
		assert.Contains(t, actualErrors, expectedMessage)
	}
}
