package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	previousEnv string
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.previousEnv = os.Getenv("GOYAVE_ENV")
}

func (suite *ConfigTestSuite) SetupTest() {
	os.Setenv("GOYAVE_ENV", "test")
	Clear()
	if err := Load(); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ConfigTestSuite) TestIsLoaded() {
	suite.True(IsLoaded())
}

func (suite *ConfigTestSuite) TestLoadDefaults() {
	// TODO test loadDefaults
}

// TODO test override

func (suite *ConfigTestSuite) TestSet() {
	suite.Equal("root level content", Get("rootLevel"))
	Set("rootLevel", "root level content override")
	suite.Equal("root level content override", Get("rootLevel"))

	suite.Equal("test", Get("app.environment"))
	Set("app.environment", "test_override")
	suite.Equal("test_override", Get("app.environment"))

	suite.Panics(func() { // empty key not allowed
		Set("", "")
	})

	// Trying to convert an entry to a category
	config["app"].(object)["category"] = object{"entry": &Entry{"value", reflect.String, []interface{}{}}}
	suite.Panics(func() {
		Set("app.category.entry.error", "override")
	})

	suite.Panics(func() {
		Set("rootLevel.error", "override")
	})

	// Trying to replace a category
	config["app"].(object)["category"] = object{"entry": &Entry{"value", reflect.String, []interface{}{}}}
	suite.Panics(func() {
		Set("app.category", "not a category")
	})
	suite.Panics(func() {
		Set("app", "not a category")
	})
}

func (suite *ConfigTestSuite) TestWalk() {
	config := object{
		"rootLevel": &Entry{"root level content", reflect.String, []interface{}{}},
		"app": object{
			"environment": &Entry{"test", reflect.String, []interface{}{}},
		},
	}
	category, entryKey, exists := walk(config, "app.environment")
	suite.True(exists)
	suite.Equal("environment", entryKey)
	suite.Equal(config["app"], category)

	category, entryKey, exists = walk(config, "category.subcategory.entry")
	suite.False(exists)
	suite.Equal("entry", entryKey)
	n, ok := config["category"]
	suite.True(ok)
	newCat, ok := n.(object)
	suite.True(ok)

	s, ok := newCat["subcategory"]
	suite.True(ok)
	subCat, ok := s.(object)
	suite.True(ok)
	suite.Equal(subCat, category)

	_, ok = subCat["entry"]
	suite.False(ok)

	category, entryKey, exists = walk(config, "category.subcategory.other")
	suite.False(exists)
	suite.Equal("other", entryKey)
	n, ok = config["category"]
	suite.True(ok)
	newCat, ok = n.(object)
	suite.True(ok)

	s, ok = newCat["subcategory"]
	suite.True(ok)
	subCat, ok = s.(object)
	suite.True(ok)
	suite.Equal(subCat, category)

	_, ok = subCat["other"]
	suite.False(ok)

	// Trying to convert an entry to a category
	suite.Panics(func() {
		walk(config, "app.environment.error")
	})
	suite.Panics(func() {
		walk(config, "rootLevel.error")
	})

	// Trying to replace a category
	suite.Panics(func() {
		walk(config, "category.subcategory")
	})
	suite.Panics(func() {
		walk(config, "app")
	})
}

func (suite *ConfigTestSuite) TestSetCreateCategories() {
	// Entirely new categories
	Set("rootCategory.subCategory.entry", "new")
	suite.Equal("new", Get("rootCategory.subCategory.entry"))
	rootCategory, ok := config["rootCategory"]
	rootCategoryObj, okTA := rootCategory.(object)
	suite.True(ok)
	suite.True(okTA)

	subCategory, ok := rootCategoryObj["subCategory"]
	subCategoryObj, okTA := subCategory.(object)
	suite.True(ok)
	suite.True(okTA)

	e, ok := subCategoryObj["entry"]
	entry, okTA := e.(*Entry)
	suite.True(ok)
	suite.True(okTA)
	suite.Equal("new", entry.Value)

	// With a category that already exists
	Set("app.subCategory.entry", "new")
	suite.Equal("new", Get("app.subCategory.entry"))
	appCategory, ok := config["app"]
	appCategoryObj, okTA := appCategory.(object)
	suite.True(ok)
	suite.True(okTA)

	subCategory, ok = appCategoryObj["subCategory"]
	subCategoryObj, okTA = subCategory.(object)
	suite.True(ok)
	suite.True(okTA)

	e, ok = subCategoryObj["entry"]
	entry, okTA = e.(*Entry)
	suite.True(ok)
	suite.True(okTA)
	suite.Equal("new", entry.Value)
}

func (suite *ConfigTestSuite) TestSetValidation() {
	// TODO implement TestSetValidation
}

func (suite *ConfigTestSuite) TestUnset() {
	suite.Equal("root level content", Get("rootLevel"))
	Set("rootLevel", nil)
	val, ok := get("rootLevel")
	suite.False(ok)
	suite.Nil(val)
}

func (suite *ConfigTestSuite) TestGet() {
	suite.Equal("goyave", Get("app.name"))
	suite.Panics(func() {
		Get("missingKey")
	})
	suite.Panics(func() {
		Get("app.missingKey")
	})

	suite.Panics(func() {
		Get("app") // Cannot get a category
	})

	suite.Panics(func() {
		Get("server.tlsCert") // Value is nil, so considered unset
	})

	suite.Equal("goyave", GetString("app.name"))
	suite.Panics(func() {
		GetString("missingKey")
	})
	suite.Panics(func() {
		GetString("app.debug") // Not a string
	})

	suite.Equal(true, GetBool("app.debug"))
	suite.Panics(func() {
		GetBool("missingKey")
	})
	suite.Panics(func() {
		GetBool("app.name") // Not a bool
	})

	suite.Equal(8080, GetInt("server.port"))
	suite.Panics(func() {
		GetInt("missingKey")
	})
	suite.Panics(func() {
		GetInt("app.name") // Not an int
	})

	Set("testFloat", 1.42)
	suite.Equal(1.42, GetFloat("testFloat"))
	suite.Panics(func() {
		GetFloat("missingKey")
	})
	suite.Panics(func() {
		GetFloat("app.name") // Not a float
	})
}

func (suite *ConfigTestSuite) TestLowLevelGet() {
	val, ok := get("rootLevel")
	suite.True(ok)
	suite.Equal("root level content", val)

	val, ok = get("app")
	suite.False(ok)
	suite.Nil(val)

	val, ok = get("app.environment")
	suite.True(ok)
	suite.Equal("test", val)

	val, ok = get("app.notakey")
	suite.False(ok)
	suite.Nil(val)

	// Existing but unset value (nil)
	val, ok = get("server.tlsCert")
	suite.False(ok)
	suite.Nil(val)

	// Ensure getting a category is not possible
	config["app"].(object)["test"] = object{"this": &Entry{"that", reflect.String, []interface{}{}}}
	val, ok = get("app.test")
	suite.False(ok)
	suite.Nil(val)

	val, ok = get("app.test.this")
	suite.True(ok)
	suite.Equal("that", val)

	// Test path ending with a dot
}

func (suite *ConfigTestSuite) TestHas() {
	suite.False(Has("not_a_config_entry"))
	suite.True(Has("rootLevel"))
	suite.True(Has("app.name"))
}

func (suite *ConfigTestSuite) TestGetEnv() {
	os.Setenv("GOYAVE_ENV", "localhost")
	suite.Equal("config.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "test")
	suite.Equal("config.test.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "production")
	suite.Equal("config.production.json", getConfigFilePath())

	os.Setenv("GOYAVE_ENV", "test")
}

func (suite *ConfigTestSuite) TestInvalidConfig() { // TODO add custom entry validation
	// val := Get("app.name")

	// TODO re-enable config validation tests
	// config["app.name"] = true
	// err := validateConfig()
	// suite.NotNil(err)
	// suite.Equal("Invalid config:\n\t- \"app.name\" type must be string", err.Error())
	// config["app.name"] = val

	// suite.Panics(func() {
	// 	Set("app.name", true)
	// })

	// val = Get("database.connection")

	// TODO re-enable config validation tests
	// config["database.connection"] = "not a driver"
	// err = validateConfig()
	// suite.NotNil(err)
	// suite.Equal("Invalid config:\n\t- \"database.connection\" must have one of the following values: none, mysql, postgres, sqlite3, mssql", err.Error())
	// config["database.connection"] = val

	// suite.Panics(func() {
	// 	Set("server.protocol", "ftp") // Unsupported protocol
	// })

	// os.Setenv("GOYAVE_ENV", "test_invalid")
	// config = nil
	// suite.NotNil(Load())
	// os.Setenv("GOYAVE_ENV", "test")
}

func (suite *ConfigTestSuite) TestTryIntConversion() {
	e := &Entry{1.42, reflect.Int, []interface{}{}}
	suite.False(e.tryIntConversion(reflect.Float64))

	e.Value = float64(2)
	suite.True(e.tryIntConversion(reflect.Float64))
	suite.Equal(2, e.Value)
}

func (suite *ConfigTestSuite) TestValidateEntryWithConversion() {
	e := &Entry{1.42, reflect.Int, []interface{}{}}
	category := object{"number": e}
	err := category.validate("")
	suite.NotNil(err)
	suite.Equal("\n\t- \"number\" type must be int", err.Error())

	e.Value = float64(2)
	err = category.validate("")
	suite.Nil(err)
	suite.Equal(2, category["number"].(*Entry).Value)
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
