package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/System-Glitch/goyave/v2/helper/filesystem"
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

func (suite *ConfigTestSuite) TestReadConfigFile() {
	obj, err := readConfigFile("config.test.json")
	suite.Nil(err)
	suite.NotEmpty(obj)
	suite.Equal("root level content", obj["rootLevel"])
	cat, ok := obj["app"]
	suite.True(ok)
	catObj, ok := cat.(map[string]interface{})
	suite.True(ok)
	suite.Equal("test", catObj["environment"])

	_, ok = catObj["name"]
	suite.False(ok)

	// Error
	err = ioutil.WriteFile("test-forbidden.json", []byte("{\"app\":\"test\"}"), 0111)
	if err != nil {
		panic(err)
	}
	defer filesystem.Delete("test-forbidden.json")
	obj, err = readConfigFile("test-forbidden.json")

	suite.NotNil(err)
	if err != nil {
		suite.Equal("open test-forbidden.json: permission denied", err.Error())
	}

	suite.Empty(obj)
}

func (suite *ConfigTestSuite) TestLoadDefaults() {
	src := object{
		"rootLevel": &Entry{"root level content", reflect.String, []interface{}{}},
		"app": object{
			"environment": &Entry{"test", reflect.String, []interface{}{}},
		},
		"auth": object{
			"basic": object{
				"username": &Entry{"test username", reflect.String, []interface{}{}},
				"password": &Entry{"test password", reflect.String, []interface{}{}},
			},
		},
	}
	dst := object{}
	loadDefaults(src, dst)

	e, ok := dst["rootLevel"]
	suite.True(ok)
	entry, ok := e.(*Entry)
	suite.True(ok)
	suite.Equal("root level content", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)
	suite.NotSame(src["rootLevel"], dst["rootLevel"])

	e, ok = dst["app"]
	suite.True(ok)
	app, ok := e.(object)
	suite.True(ok)

	e, ok = app["environment"]
	suite.True(ok)
	suite.Equal("test", e.(*Entry).Value)

	e, ok = dst["auth"]
	suite.True(ok)
	auth, ok := e.(object)
	suite.True(ok)

	e, ok = auth["basic"]
	suite.True(ok)
	basic, ok := e.(object)
	suite.True(ok)

	e, ok = basic["username"]
	suite.True(ok)
	entry, ok = e.(*Entry)
	suite.True(ok)
	suite.Equal("test username", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)

	e, ok = basic["password"]
	suite.True(ok)
	entry, ok = e.(*Entry)
	suite.True(ok)
	suite.Equal("test password", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)
}

func (suite *ConfigTestSuite) TestOverride() {
	src := object{
		"rootLevel": "root level content",
		"app": map[string]interface{}{
			"environment": "test",
		},
		"auth": map[string]interface{}{
			"basic": map[string]interface{}{
				"username": "test username",
				"password": "test password",
				"deepcategory": map[string]interface{}{
					"deepentry": 1,
				},
			},
		},
	}
	dst := object{
		"app": object{
			"name":        &Entry{"default name", reflect.String, []interface{}{}},
			"environment": &Entry{"default env", reflect.String, []interface{}{}},
		},
	}
	suite.Nil(override(src, dst))

	e, ok := dst["rootLevel"]
	suite.True(ok)
	entry, ok := e.(*Entry)
	suite.True(ok)
	suite.Equal("root level content", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)

	e, ok = dst["app"]
	suite.True(ok)
	app, ok := e.(object)
	suite.True(ok)

	e, ok = app["name"]
	suite.True(ok)
	suite.Equal("default name", e.(*Entry).Value)

	e, ok = app["environment"]
	suite.True(ok)
	suite.Equal("test", e.(*Entry).Value)

	e, ok = dst["auth"]
	suite.True(ok)
	auth, ok := e.(object)
	suite.True(ok)

	e, ok = auth["basic"]
	suite.True(ok)
	basic, ok := e.(object)
	suite.True(ok)

	e, ok = basic["username"]
	suite.True(ok)
	entry, ok = e.(*Entry)
	suite.True(ok)
	suite.Equal("test username", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)

	e, ok = basic["password"]
	suite.True(ok)
	entry, ok = e.(*Entry)
	suite.True(ok)
	suite.Equal("test password", entry.Value)
	suite.Equal(reflect.String, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)

	e, ok = basic["deepcategory"]
	suite.True(ok)
	deepCategory, ok := e.(object)
	suite.True(ok)

	e, ok = deepCategory["deepentry"]
	suite.True(ok)
	entry, ok = e.(*Entry)
	suite.True(ok)
	suite.Equal(1, entry.Value)
	suite.Equal(reflect.Int, entry.Type)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)
}

func (suite *ConfigTestSuite) TestOverrideConflict() {
	// conflict override entry with category (depth == 0)
	src := object{
		"rootLevel": map[string]interface{}{
			"environment": "test",
		},
	}
	dst := object{
		"rootLevel": &Entry{"root level content", reflect.String, []interface{}{}},
	}
	err := override(src, dst)
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override entry \"rootLevel\" with a category", err.Error())
	}

	// conflict override entry with category (depth > 0)
	src = object{
		"app": map[string]interface{}{
			"environment": map[string]interface{}{
				"prod": false,
			},
		},
	}
	dst = object{
		"app": object{
			"environment": &Entry{"default env", reflect.String, []interface{}{}},
		},
	}
	err = override(src, dst)
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override entry \"environment\" with a category", err.Error())
	}

	// conflict override category with entry (depth == 0)
	src = object{
		"app": "test",
	}
	dst = object{
		"app": object{
			"name":        &Entry{"default name", reflect.String, []interface{}{}},
			"environment": &Entry{"default env", reflect.String, []interface{}{}},
		},
	}
	err = override(src, dst)
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override category \"app\" with an entry", err.Error())
	}

	// conflict override category with entry (depth > 0)
	src = object{
		"app": map[string]interface{}{
			"environments": "test",
		},
	}
	dst = object{
		"app": object{
			"name": &Entry{"default name", reflect.String, []interface{}{}},
			"environments": object{
				"prod": &Entry{false, reflect.Bool, []interface{}{}},
			},
		},
	}
	err = override(src, dst)
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override category \"environments\" with an entry", err.Error())
	}
}

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

	// Path ending with a dot
	suite.Panics(func() {
		walk(config, "paniccategory.subcategory.")
	})
	// Check nothing has been created
	_, ok = config["paniccategory"]
	suite.False(ok)

	suite.Panics(func() { // empty key not allowed
		walk(config, "")
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

	// Path ending with a dot
	val, ok = get("app.test.")
	suite.False(ok)
	suite.Nil(val)
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

func (suite *ConfigTestSuite) TestValidateEntry() {
	// Unset (no validation needed)
	e := &Entry{nil, reflect.String, []interface{}{}}
	err := e.validate("entry")
	suite.Nil(err)

	e = &Entry{nil, reflect.String, []interface{}{"val1", "val2"}}
	err = e.validate("entry")
	suite.Nil(err)

	// Wrong type
	e = &Entry{1, reflect.String, []interface{}{}}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" type must be string", err.Error())
	}

	// Int conversion
	e = &Entry{1.0, reflect.Int, []interface{}{}}
	err = e.validate("entry")
	suite.Nil(err)
	suite.Equal(1, e.Value)

	e = &Entry{1.42, reflect.Int, []interface{}{}}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" type must be int", err.Error())
	}

	// Authorized values
	e = &Entry{1.42, reflect.Float64, []interface{}{1.2, 1.3, 2.4, 42.1, 1.4200000001}}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" must have one of the following values: [1.2 1.3 2.4 42.1 1.4200000001]", err.Error())
	}

	e = &Entry{"test", reflect.String, []interface{}{"val1", "val2"}}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" must have one of the following values: [val1 val2]", err.Error())
	}

	// Everything's fine
	e = &Entry{"val1", reflect.String, []interface{}{"val1", "val2"}}
	err = e.validate("entry")
	suite.Nil(err)

	e = &Entry{1.42, reflect.Float64, []interface{}{1.2, 1.3, 2.4, 42.1, 1.4200000001, 1.42}}
	err = e.validate("entry")
	suite.Nil(err)
}

func (suite *ConfigTestSuite) TestValidateObject() {
	config := object{
		"rootLevel": &Entry{"root level content", reflect.Bool, []interface{}{}},
		"app": object{
			"environment": &Entry{true, reflect.String, []interface{}{}},
			"subcategory": object{
				"entry": &Entry{666, reflect.Int, []interface{}{1, 2, 3}},
			},
		},
	}

	err := config.validate("")
	suite.NotNil(err)
	if err != nil {
		message := []string{"\n\t- \"rootLevel\" type must be bool",
			"\n\t- \"app.environment\" type must be string",
			"\n\t- \"app.subcategory.entry\" must have one of the following values: [1 2 3]"}
		// Maps are unordered, use a slice to make sure all messages are there
		for _, m := range message {
			suite.Contains(err.Error(), m)
		}
	}

	config = object{
		"rootLevel": &Entry{"root level content", reflect.String, []interface{}{}},
		"app": object{
			"environment": &Entry{"local", reflect.String, []interface{}{}},
			"subcategory": object{
				"entry": &Entry{2, reflect.Int, []interface{}{1, 2, 3}},
			},
		},
	}
	err = config.validate("")
	suite.Nil(err)
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
