package lang

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/fsutil/osfs"
)

type LangTestSuite struct {
	suite.Suite
}

func setRootWorkingDirectory() {
	_, filename, _, _ := runtime.Caller(1)
	directory := path.Dir(filename)
	for !fsutil.FileExists(&osfs.FS{}, path.Join(directory, "go.mod")) {
		directory = path.Join(directory, "..")
		if !fsutil.IsDirectory(&osfs.FS{}, directory) {
			panic("Couldn't find project's root directory.")
		}
	}
	if err := os.Chdir(directory); err != nil {
		panic(err)
	}
}

func (suite *LangTestSuite) SetupSuite() {
	setRootWorkingDirectory()
}

func (suite *LangTestSuite) TestNew() {
	l := New()
	expected := &Languages{
		languages: map[string]*Language{enUS.name: enUS},
		Default:   enUS.name,
	}
	suite.Equal(expected, l)
}

func (suite *LangTestSuite) TestLoadError() {
	l := New()
	err := l.Load(&osfs.FS{}, "notalanguagedir", "notalanguagepath")
	suite.Error(err)
	suite.Len(l.languages, 1)
}

func (suite *LangTestSuite) TestLoadInvalid() {
	dst := map[string]string{}
	suite.Error(readLangFile(&osfs.FS{}, "resources/lang/invalid.json", &dst))
}

func (suite *LangTestSuite) TestLoadOverride() {
	l := New()
	err := l.Load(&osfs.FS{}, "en-US", "resources/lang/en-US")
	suite.NoError(err)
	suite.Len(l.languages, 1)
	suite.Equal("rule override", l.languages["en-US"].validation.rules["override"])
	suite.Equal("Custom line", l.languages["en-US"].lines["custom-line"])
}

func (suite *LangTestSuite) TestLoad() {
	l := New()
	err := l.Load(&osfs.FS{}, "en-UK", "resources/lang/en-UK")
	suite.NoError(err)
	suite.Len(l.languages, 2)
	expected := &Language{
		name: "en-UK",
		lines: map[string]string{
			"malformed-request": "Malformed request",
			"malformed-json":    "Malformed JSON",
			"test-load":         "load UK",
		},
		validation: validationLines{
			rules:  map[string]string{},
			fields: map[string]string{},
		},
	}
	suite.Equal(expected, l.languages["en-UK"])

	err = l.Load(&osfs.FS{}, "en-UK", "resources/lang/en-US") // Overriding en-UK with the lines in en-US
	suite.NoError(err)
	suite.Len(l.languages, 2)
	expected = &Language{
		name: "en-UK",
		lines: map[string]string{
			"malformed-request": "Malformed request",
			"malformed-json":    "Malformed JSON",
			"custom-line":       "Custom line",
			"placeholder":       "Line with :placeholders",
			"many-placeholders": "Line with :count :placeholders",
			"test-load":         "load US",
		},
		validation: validationLines{
			rules: map[string]string{
				"override":       "rule override",
				"required.array": "The :field values are required.",
			},
			fields: map[string]string{
				"email": "email address",
			},
		},
	}
	suite.Equal(expected, l.languages["en-UK"])
}

func (suite *LangTestSuite) TestLoadAllAvailableLanguages() {
	l := New()
	suite.NoError(l.LoadAllAvailableLanguages(&osfs.FS{}))
	suite.Len(l.languages, 2)
	suite.Contains(l.languages, "en-US")
	suite.Contains(l.languages, "en-UK")
}

func (suite *LangTestSuite) TestGetLanguage() {
	l := New()
	if err := l.LoadAllAvailableLanguages(&osfs.FS{}); err != nil {
		suite.Error(err)
		return
	}
	suite.Equal(l.languages["en-UK"], l.GetLanguage("en-UK"))
	suite.Equal("dummy", l.GetLanguage("fr-FR").Name())
}

func (suite *LangTestSuite) TestDummyLanguage() {
	l := New()
	lang := l.GetLanguage("dummy")
	suite.Equal("malformed-request", lang.Get("malformed-request"))
	suite.Equal("validation.rules.required", lang.Get("validation.rules.required"))
	suite.Equal("validation.fields.email", lang.Get("validation.fields.email"))
}

func (suite *LangTestSuite) TestGetDefault() {
	l := New()
	suite.Equal(l.languages["en-US"], l.GetDefault())
}

func (suite *LangTestSuite) TestIsAvailable() {
	l := New()
	suite.True(l.IsAvailable("en-US"))
	suite.False(l.IsAvailable("fr-FR"))
}

func (suite *LangTestSuite) TestGetAvailableLanguages() {
	l := New()
	if err := l.LoadAllAvailableLanguages(&osfs.FS{}); err != nil {
		suite.Error(err)
		return
	}

	suite.ElementsMatch([]string{"en-US", "en-UK"}, l.GetAvailableLanguages())
}

func (suite *LangTestSuite) TestDetectLanguage() {
	l := New()
	if err := l.Load(&osfs.FS{}, "fr-FR", "resources/lang/en-US"); err != nil {
		panic(err)
	}

	suite.Equal(l.languages["en-US"], l.DetectLanguage("en"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage("en-US, fr"))
	suite.Equal(l.languages["fr-FR"], l.DetectLanguage("fr-FR"))
	suite.Equal(l.languages["fr-FR"], l.DetectLanguage("fr"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage("fr, en-US"))
	suite.Equal(l.languages["fr-FR"], l.DetectLanguage("fr-FR, en-US"))
	suite.Equal(l.languages["fr-FR"], l.DetectLanguage("fr, en-US;q=0.9"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage("en, fr-FR;q=0.9"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage("*"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage("notalang"))
	suite.Equal(l.languages["en-US"], l.DetectLanguage(""))
}

func (suite *LangTestSuite) TestLanguagesGet() {
	l := New()
	if err := l.LoadAllAvailableLanguages(&osfs.FS{}); err != nil {
		suite.Error(err)
		return
	}

	suite.Equal("malformed-request", l.Get("fr-FR", "malformed-request"))
	suite.Equal("Malformed request", l.Get("en-US", "malformed-request"))
	suite.Equal("Line with awesomeness", l.Get("en-US", "placeholder", ":placeholders", "awesomeness"))
}

func (suite *LangTestSuite) TestGet() {
	l := New()
	if err := l.LoadAllAvailableLanguages(&osfs.FS{}); err != nil {
		suite.Error(err)
		return
	}

	lang := l.GetLanguage("en-US")
	suite.Equal("Malformed request", lang.Get("malformed-request"))
	suite.Equal("notaline", lang.Get("notaline"))
	suite.Equal("validation.rules.notarule", lang.Get("validation.rules.notarule"))
	suite.Equal("validation.rules.notafield", lang.Get("validation.rules.notafield"))
	suite.Equal("Line with an infinite amount of awesomeness", lang.Get("many-placeholders", ":placeholders", "awesomeness", ":count", "an infinite amount of"))
}

func (suite *LangTestSuite) TestMerge() {
	dst := &Language{
		lines: map[string]string{"line": "line 1"},
		validation: validationLines{
			rules: map[string]string{},
			fields: map[string]string{
				"test": "test field",
			},
		},
	}
	src := &Language{
		lines: map[string]string{"other": "line 2"},
		validation: validationLines{
			rules: map[string]string{},
			fields: map[string]string{
				"email": "email address",
				"test":  "test field override",
			},
		},
	}

	mergeLang(dst, src)
	suite.Equal("line 1", dst.lines["line"])
	suite.Equal("line 2", dst.lines["other"])

	suite.Equal("email address", dst.validation.fields["email"])

	suite.Equal("test field override", dst.validation.fields["test"])
}

func (suite *LangTestSuite) TestPlaceholders() {
	suite.Equal("notaline", convertEmptyLine("notaline", "", []string{":username", "Kevin"}))
	suite.Equal("Greetings, Kevin", convertEmptyLine("greetings", "Greetings, :username", []string{":username", "Kevin"}))
	suite.Equal("Greetings, Kevin, today is Monday", convertEmptyLine("greetings", "Greetings, :username, today is :today", []string{":username", "Kevin", ":today", "Monday"}))
	suite.Equal("Greetings, Kevin, today is :today", convertEmptyLine("greetings", "Greetings, :username, today is :today", []string{":username", "Kevin", ":today"}))
}

func (suite *LangTestSuite) TestSetDefault() {
	SetDefaultLine("test-line", "It's sunny today")
	suite.Equal("It's sunny today", enUS.lines["test-line"])
	delete(enUS.lines, "test-line")

	SetDefaultValidationRule("test-validation-rules", "It's sunny today")
	suite.Equal("It's sunny today", enUS.validation.rules["test-validation-rules"])
	delete(enUS.validation.rules, "test-validation-rules")

	SetDefaultFieldName("test-field-name", "Sun")
	suite.Equal("Sun", enUS.validation.fields["test-field-name"])
	delete(enUS.validation.fields, "test-field-name")

	// Test no override
	enUS.validation.fields["test-field"] = "test"
	SetDefaultFieldName("test-field", "Sun")
	suite.Equal("Sun", enUS.validation.fields["test-field"])
	delete(enUS.validation.fields, "test-field")
}

func TestLangTestSuite(t *testing.T) {
	suite.Run(t, new(LangTestSuite))
}
