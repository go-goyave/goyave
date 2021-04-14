package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v3/helper/filesystem"
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
		"rootLevel": &Entry{"root level content", []interface{}{}, reflect.String, false},
		"app": object{
			"environment": &Entry{"test", []interface{}{}, reflect.String, false},
		},
		"auth": object{
			"basic": object{
				"username": &Entry{"test username", []interface{}{}, reflect.String, false},
				"password": &Entry{"test password", []interface{}{}, reflect.String, false},
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

func (suite *ConfigTestSuite) TestLoadDefaultsWithSlice() {
	slice := []string{"val1", "val2"}
	src := object{
		"rootLevel": &Entry{slice, []interface{}{}, reflect.String, true},
	}
	dst := object{}
	loadDefaults(src, dst)

	e, ok := dst["rootLevel"]
	suite.True(ok)
	entry, ok := e.(*Entry)
	suite.True(ok)

	cpy, ok := entry.Value.([]string)
	suite.True(ok)
	suite.NotSame(cpy, slice)
	suite.Equal(slice, cpy)
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
			"name":        &Entry{"default name", []interface{}{}, reflect.String, false},
			"environment": &Entry{"default env", []interface{}{}, reflect.String, false},
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
		"rootLevel": &Entry{"root level content", []interface{}{}, reflect.String, false},
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
			"environment": &Entry{"default env", []interface{}{}, reflect.String, false},
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
			"name":        &Entry{"default name", []interface{}{}, reflect.String, false},
			"environment": &Entry{"default env", []interface{}{}, reflect.String, false},
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
			"name": &Entry{"default name", []interface{}{}, reflect.String, false},
			"environments": object{
				"prod": &Entry{false, []interface{}{}, reflect.Bool, false},
			},
		},
	}
	err = override(src, dst)
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override category \"environments\" with an entry", err.Error())
	}
}

func (suite *ConfigTestSuite) TestLoad() {
	Clear()
	err := Load()
	suite.Nil(err)
	suite.Equal(configDefaults["server"], config["server"])
	suite.Equal(configDefaults["database"], config["database"])
	suite.NotEqual(configDefaults["app"], config["app"])

	defaultAppCategory := configDefaults["app"].(object)
	appCategory := config["app"].(object)
	suite.Equal(defaultAppCategory["name"], appCategory["name"])
	suite.Equal(defaultAppCategory["debug"], appCategory["debug"])
	suite.Equal(defaultAppCategory["defaultLanguage"], appCategory["defaultLanguage"])
	suite.Equal(defaultAppCategory["name"], appCategory["name"])

	// readConfigFile error
	Clear()
	err = ioutil.WriteFile("config.forbidden.json", []byte("{\"app\":\"test\"}"), 0111)
	if err != nil {
		panic(err)
	}
	defer filesystem.Delete("config.forbidden.json")
	if e := os.Setenv("GOYAVE_ENV", "forbidden"); e != nil {
		panic(e)
	}
	err = Load()
	if e := os.Setenv("GOYAVE_ENV", "test"); e != nil {
		panic(e)
	}
	suite.NotNil(err)
	if err != nil {
		suite.Equal("open config.forbidden.json: permission denied", err.Error())
	}
	suite.Nil(config)
	suite.False(IsLoaded())

	// override error
	Clear()
	configDefaults["rootLevel"] = object{}
	err = Load()
	delete(configDefaults, "rootLevel")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- Cannot override category \"rootLevel\" with an entry", err.Error())
	}
	suite.Nil(config)
	suite.False(IsLoaded())

	// validation error
	Clear()
	configDefaults["rootLevel"] = &Entry{42, []interface{}{}, reflect.Int, false}
	err = Load()
	delete(configDefaults, "rootLevel")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("Invalid config:\n\t- \"rootLevel\" type must be int", err.Error())
	}
	suite.Nil(config)
	suite.False(IsLoaded())
}

func (suite *ConfigTestSuite) TestLoadFrom() {
	Clear()
	err := LoadFrom("../resources/custom_config.json")
	suite.Nil(err)
	suite.Equal(configDefaults["server"], config["server"])
	suite.Equal(configDefaults["database"], config["database"])
	suite.Equal(configDefaults["app"], config["app"])

	e, ok := config["custom-entry"]
	suite.True(ok)
	entry, ok := e.(*Entry)
	suite.True(ok)
	suite.Equal("value", entry.Value)
}

func (suite *ConfigTestSuite) TestCreateMissingCategories() {
	config := object{}
	created := createMissingCategories(config, "category.entry")
	e, ok := config["category"]
	suite.True(ok)
	category, ok := e.(object)
	suite.True(ok)

	_, ok = category["entry"]
	suite.False(ok)
	suite.Equal(category, created)

	// Depth
	config = object{}
	created = createMissingCategories(config, "category.subcategory.entry")
	e, ok = config["category"]
	suite.True(ok)
	category, ok = e.(object)
	suite.True(ok)

	e, ok = category["subcategory"]
	suite.True(ok)
	subcategory, ok := e.(object)
	suite.True(ok)

	_, ok = subcategory["entry"]
	suite.False(ok)
	suite.Equal(subcategory, created)

	// With partial existence
	config = object{
		"category": object{},
	}
	created = createMissingCategories(config, "category.subcategory.entry")
	e, ok = config["category"]
	suite.True(ok)
	category, ok = e.(object)
	suite.True(ok)

	e, ok = category["subcategory"]
	suite.True(ok)
	subcategory, ok = e.(object)
	suite.True(ok)

	_, ok = subcategory["entry"]
	suite.False(ok)
	suite.Equal(subcategory, created)

	config = object{}
	created = createMissingCategories(config, "entry")
	suite.Equal(config, created)

	_, ok = config["entry"]
	suite.False(ok)
}

func (suite *ConfigTestSuite) TestSet() {
	suite.Equal("root level content", Get("rootLevel"))
	Set("rootLevel", "root level content override")
	suite.Equal("root level content override", Get("rootLevel"))

	suite.Equal("test", Get("app.environment"))
	Set("app.environment", "test_override")
	suite.Equal("test_override", Get("app.environment"))

	Set("newEntry", "test_new_entry")
	e, ok := config["newEntry"]
	suite.True(ok)
	entry, ok := e.(*Entry)
	suite.True(ok)
	if ok {
		suite.Equal("test_new_entry", entry.Value)
		suite.Equal(entry.Type, reflect.String)
		suite.False(entry.IsSlice)
		suite.Equal([]interface{}{}, entry.AuthorizedValues)
	}

	suite.Panics(func() { // empty key not allowed
		Set("", "")
	})

	// Slice
	config["stringslice"] = &Entry{nil, []interface{}{}, reflect.String, true}
	Set("stringslice", []string{"val1", "val2"})
	suite.Equal([]string{"val1", "val2"}, config["stringslice"].(*Entry).Value)

	// Trying to convert an entry to a category
	config["app"].(object)["category"] = object{"entry": &Entry{"value", []interface{}{}, reflect.String, false}}
	suite.Panics(func() {
		Set("app.category.entry.error", "override")
	})

	suite.Panics(func() {
		Set("rootLevel.error", "override")
	})

	// Trying to replace a category
	config["app"].(object)["category"] = object{"entry": &Entry{"value", []interface{}{}, reflect.String, false}}
	suite.Panics(func() {
		Set("app.category", "not a category")
	})
	suite.Panics(func() {
		Set("app", "not a category")
	})

	// Config not loaded
	Clear()
	suite.Panics(func() {
		Set("app.name", "config not loaded")
	})
}

func (suite *ConfigTestSuite) TestWalk() {
	config := object{
		"rootLevel": &Entry{"root level content", []interface{}{}, reflect.String, false},
		"app": object{
			"environment": &Entry{"test", []interface{}{}, reflect.String, false},
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
	previous := Get("app.name")
	suite.Panics(func() {
		Set("app.name", 1)
	})
	suite.Equal(previous, Get("app.name"))

	// Works with float64 without decimals
	Set("server.port", 8080.0)
	suite.Equal(8080, Get("server.port"))
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
		Get("server.tls.cert") // Value is nil, so considered unset
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

func (suite *ConfigTestSuite) TestGetSlice() {
	Set("stringslice", []string{"val1", "val2"})
	suite.Equal([]string{"val1", "val2"}, GetStringSlice("stringslice"))
	suite.Panics(func() {
		GetStringSlice("missingKey")
	})
	suite.Panics(func() {
		GetStringSlice("app.name") // Not a string slice
	})

	Set("boolslice", []bool{true, false})
	suite.Equal([]bool{true, false}, GetBoolSlice("boolslice"))
	suite.Panics(func() {
		GetBoolSlice("missingKey")
	})
	suite.Panics(func() {
		GetBoolSlice("app.name") // Not a bool slice
	})

	Set("intslice", []int{1, 2})
	suite.Equal([]int{1, 2}, GetIntSlice("intslice"))
	suite.Panics(func() {
		GetIntSlice("missingKey")
	})
	suite.Panics(func() {
		GetIntSlice("app.name") // Not an int slice
	})

	Set("floatslice", []float64{1.2, 2.3})
	suite.Equal([]float64{1.2, 2.3}, GetFloatSlice("floatslice"))
	suite.Panics(func() {
		GetFloatSlice("missingKey")
	})
	suite.Panics(func() {
		GetFloatSlice("app.name") // Not a float slice
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
	val, ok = get("server.tls.cert")
	suite.False(ok)
	suite.Nil(val)

	// Ensure getting a category is not possible
	config["app"].(object)["test"] = object{"this": &Entry{"that", []interface{}{}, reflect.String, false}}
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

	// Config not loaded
	Clear()
	suite.Panics(func() {
		get("app.name")
	})
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
	e := &Entry{1.42, []interface{}{}, reflect.Int, false}
	suite.False(e.tryIntConversion(reflect.Float64))

	e.Value = float64(2)
	suite.True(e.tryIntConversion(reflect.Float64))
	suite.Equal(2, e.Value)
}

func (suite *ConfigTestSuite) TestValidateEntryWithConversion() {
	e := &Entry{1.42, []interface{}{}, reflect.Int, false}
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
	e := &Entry{nil, []interface{}{}, reflect.String, false}
	err := e.validate("entry")
	suite.Nil(err)

	e = &Entry{nil, []interface{}{"val1", "val2"}, reflect.String, false}
	err = e.validate("entry")
	suite.Nil(err)

	// Wrong type
	e = &Entry{1, []interface{}{}, reflect.String, false}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" type must be string", err.Error())
	}

	// Int conversion
	e = &Entry{1.0, []interface{}{}, reflect.Int, false}
	err = e.validate("entry")
	suite.Nil(err)
	suite.Equal(1, e.Value)

	e = &Entry{1.42, []interface{}{}, reflect.Int, false}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" type must be int", err.Error())
	}

	// Authorized values
	e = &Entry{1.42, []interface{}{1.2, 1.3, 2.4, 42.1, 1.4200000001}, reflect.Float64, false}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" must have one of the following values: [1.2 1.3 2.4 42.1 1.4200000001]", err.Error())
	}

	e = &Entry{"test", []interface{}{"val1", "val2"}, reflect.String, false}
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" must have one of the following values: [val1 val2]", err.Error())
	}

	// Everything's fine
	e = &Entry{"val1", []interface{}{"val1", "val2"}, reflect.String, false}
	err = e.validate("entry")
	suite.Nil(err)

	e = &Entry{1.42, []interface{}{1.2, 1.3, 2.4, 42.1, 1.4200000001, 1.42}, reflect.Float64, false}
	err = e.validate("entry")
	suite.Nil(err)

	// From environment variable
	e = &Entry{"${TEST_VAR}", []interface{}{}, reflect.Float64, false}
	os.Setenv("TEST_VAR", "2..")
	defer os.Unsetenv("TEST_VAR")
	err = e.validate("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" could not be converted to float64 from environment variable \"TEST_VAR\" of value \"2..\"", err.Error())
	}
	suite.Equal("${TEST_VAR}", e.Value)
}

func (suite *ConfigTestSuite) TestValidateObject() {
	config := object{
		"rootLevel": &Entry{"root level content", []interface{}{}, reflect.Bool, false},
		"app": object{
			"environment": &Entry{true, []interface{}{}, reflect.String, false},
			"subcategory": object{
				"entry": &Entry{666, []interface{}{1, 2, 3}, reflect.Int, false},
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
		"rootLevel": &Entry{"root level content", []interface{}{}, reflect.String, false},
		"app": object{
			"environment": &Entry{"local", []interface{}{}, reflect.String, false},
			"subcategory": object{
				"entry": &Entry{2, []interface{}{1, 2, 3}, reflect.Int, false},
			},
		},
	}
	err = config.validate("")
	suite.Nil(err)
}

func (suite *ConfigTestSuite) TestRegister() {
	entry := Entry{"value", []interface{}{"value", "other value"}, reflect.String, false}
	Register("rootLevel", entry)
	newEntry, ok := configDefaults["rootLevel"]
	suite.True(ok)
	suite.Equal(&entry, newEntry)
	suite.NotSame(&entry, newEntry)
	delete(configDefaults, "rootLevel")

	// Entry already exists and matches -> do nothing
	appCategory := configDefaults["app"].(object)
	entry = Entry{"goyave", []interface{}{}, reflect.String, false}
	current := appCategory["name"]
	Register("app.name", entry)
	newEntry, ok = appCategory["name"]
	suite.True(ok)
	suite.Same(current, newEntry)
	suite.NotSame(&entry, newEntry)

	// Entry already exists but doesn't match -> panic

	// Value doesn't match
	entry = Entry{"not goyave", []interface{}{}, reflect.String, false}
	current = appCategory["name"]
	suite.Panics(func() {
		Register("app.name", entry)
	})
	newEntry, ok = appCategory["name"]
	suite.True(ok)
	suite.Same(current, newEntry)
	suite.Equal("goyave", newEntry.(*Entry).Value)

	// Type doesn't match
	entry = Entry{"goyave", []interface{}{}, reflect.Int, false}
	current = appCategory["name"]
	suite.Panics(func() {
		Register("app.name", entry)
	})
	newEntry, ok = appCategory["name"]
	suite.True(ok)
	suite.Same(current, newEntry)
	suite.Equal(reflect.String, newEntry.(*Entry).Type)

	// Required values don't match
	entry = Entry{"goyave", []interface{}{"app", "thing"}, reflect.String, false}
	current = appCategory["name"]
	suite.Panics(func() {
		Register("app.name", entry)
	})
	newEntry, ok = appCategory["name"]
	suite.True(ok)
	suite.Same(current, newEntry)
	suite.Equal([]interface{}{}, newEntry.(*Entry).AuthorizedValues)
}

func (suite *ConfigTestSuite) TestTryEnvVarConversion() {
	entry := &Entry{"${TEST_VAR}", []interface{}{}, reflect.String, false}
	err := entry.tryEnvVarConversion("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\": \"TEST_VAR\" environment variable is not set", err.Error())
	}
	suite.Equal("${TEST_VAR}", entry.Value)

	os.Setenv("TEST_VAR", "")
	defer os.Unsetenv("TEST_VAR")
	err = entry.tryEnvVarConversion("entry")
	suite.Nil(err)
	suite.Equal("", entry.Value)
	entry.Value = "${TEST_VAR}"

	os.Setenv("TEST_VAR", "env var value")
	err = entry.tryEnvVarConversion("entry")
	suite.Nil(err)
	suite.Equal("env var value", entry.Value)

	// Int conversion
	entry = &Entry{"${TEST_VAR}", []interface{}{}, reflect.Int, false}
	os.Setenv("TEST_VAR", "29")
	err = entry.tryEnvVarConversion("entry")
	suite.Nil(err)
	suite.Equal(29, entry.Value)
	entry.Value = "${TEST_VAR}"

	os.Setenv("TEST_VAR", "2.9")
	err = entry.tryEnvVarConversion("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" could not be converted to int from environment variable \"TEST_VAR\" of value \"2.9\"", err.Error())
	}
	suite.Equal("${TEST_VAR}", entry.Value)

	// Float conversion
	entry = &Entry{"${TEST_VAR}", []interface{}{}, reflect.Float64, false}
	os.Setenv("TEST_VAR", "2.9")
	err = entry.tryEnvVarConversion("entry")
	suite.Nil(err)
	suite.Equal(2.9, entry.Value)
	entry.Value = "${TEST_VAR}"

	os.Setenv("TEST_VAR", "2..")
	err = entry.tryEnvVarConversion("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" could not be converted to float64 from environment variable \"TEST_VAR\" of value \"2..\"", err.Error())
	}
	suite.Equal("${TEST_VAR}", entry.Value)

	// Bool conversion
	entry = &Entry{"${TEST_VAR}", []interface{}{}, reflect.Bool, false}
	os.Setenv("TEST_VAR", "true")
	err = entry.tryEnvVarConversion("entry")
	suite.Nil(err)
	suite.Equal(true, entry.Value)
	entry.Value = "${TEST_VAR}"

	os.Setenv("TEST_VAR", "no")
	err = entry.tryEnvVarConversion("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\" could not be converted to bool from environment variable \"TEST_VAR\" of value \"no\"", err.Error())
	}
	suite.Equal("${TEST_VAR}", entry.Value)

	// Empty name edge case
	entry = &Entry{"${}", []interface{}{}, reflect.Bool, false}
	err = entry.tryEnvVarConversion("entry")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"entry\": \"\" environment variable is not set", err.Error())
	}
	suite.Equal("${}", entry.Value)
}

func (suite *ConfigTestSuite) TestSlice() {
	entry := Entry{[]string{"val1", "val2"}, []interface{}{}, reflect.String, false}
	suite.NotNil(entry.validate("slice"))

	entry = Entry{[]string{"val1", "val2"}, []interface{}{}, reflect.String, true}
	suite.Nil(entry.validate("slice"))

	entry.Value = []int{4, 5}
	err := entry.validate("slice")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"slice\" must be a slice of string", err.Error())
	}

	entry = Entry{[]interface{}{"val1", 1, 2.3}, []interface{}{}, reflect.Interface, true}
	suite.Nil(entry.validate("slice"))

	entry = Entry{[]interface{}{"val1", 1, 2.3}, []interface{}{"val1", 1, 2.3, true}, reflect.Interface, true}
	suite.Nil(entry.validate("slice"))

	entry = Entry{[]interface{}{"val1", 1, 'c'}, []interface{}{"val1", 1, 2.3, true}, reflect.Interface, true}
	err = entry.validate("slice")
	suite.NotNil(err)
	if err != nil {
		suite.Equal("\"slice\" elements must have one of the following values: [val1 1 2.3 true]", err.Error())
	}
}

func (suite *ConfigTestSuite) TestSliceIntConversion() {
	entry := Entry{[]float64{1, 2}, []interface{}{}, reflect.Int, true}
	suite.Nil(entry.validate("slice"))

	suite.Equal([]int{1, 2}, entry.Value)

	entry = Entry{[]float64{1, 2.5}, []interface{}{}, reflect.Int, true}
	suite.NotNil(entry.validate("slice"))

	suite.Equal([]float64{1, 2.5}, entry.Value)
}

func (suite *ConfigTestSuite) TestMakeEntryFromValue() {
	entry := makeEntryFromValue(1)
	suite.Equal(1, entry.Value)
	suite.Equal(entry.Type, reflect.Int)
	suite.False(entry.IsSlice)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)

	entry = makeEntryFromValue([]string{"1", "2"})
	suite.Equal([]string{"1", "2"}, entry.Value)
	suite.Equal(entry.Type, reflect.String)
	suite.True(entry.IsSlice)
	suite.Equal([]interface{}{}, entry.AuthorizedValues)
}

func (suite *ConfigTestSuite) TestLoadJSON() {
	json := `
	{
		"app": {
			"name": "loaded from json"
		}
	}`

	suite.Nil(LoadJSON(json))
	suite.Equal("loaded from json", Get("app.name"))

	Clear()

	json = `
	{
		"app": {
			"name": 4
		}
	}`

	err := LoadJSON(json)
	suite.NotNil(err)
	suite.Contains(err.Error(), "Invalid config")

	json = `{`

	err = LoadJSON(json)
	suite.NotNil(err)
	suite.Contains(err.Error(), "EOF")
}

func (suite *ConfigTestSuite) TearDownAllSuite() {
	config = map[string]interface{}{}
	os.Setenv("GOYAVE_ENV", suite.previousEnv)
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
