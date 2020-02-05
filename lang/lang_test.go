package lang

import (
	"path"
	"runtime"
	"testing"

	"github.com/System-Glitch/goyave/v2/config"
	"github.com/stretchr/testify/suite"
)

type LangTestSuite struct {
	suite.Suite
}

func loadTestLang(lang string) {
	_, filename, _, _ := runtime.Caller(1)
	load(lang, path.Dir(filename)+"/../resources/lang/en-US")
}

func (suite *LangTestSuite) SetupSuite() {
	LoadDefault()
	LoadAllAvailableLanguages()
	config.Load()
	config.Set("defaultLanguage", "en-US")
}

func (suite *LangTestSuite) TestLang() {
	suite.Equal("email address", Get("en-US", "validation.fields.email"))
	suite.Equal("The :field is required.", Get("en-US", "validation.rules.required"))
	suite.Equal("Non-validated fields are forbidden.", Get("en-US", "disallow-non-validated-fields"))
	suite.Equal("These credentials don't match our records.", Get("en-US", "auth.invalid-credentials"))
	suite.Equal("doesn't.exist", Get("en-US", "doesn't.exist"))
	suite.Equal("doesn'texist", Get("en-US", "doesn'texist"))
	suite.Equal("validation.doesn't.exist", Get("en-US", "validation.doesn't.exist"))
	suite.Equal("validation.rules", Get("en-US", "validation.rules"))
	suite.Equal("validation.rules.doesn't.exist", Get("en-US", "validation.rules.doesn't.exist"))
	suite.Equal("validation.fields.doesn't", Get("en-US", "validation.fields.doesn't"))
	suite.Equal("validation.fields.doesn.t.", Get("en-US", "validation.fields.doesn.t."))

	languages["en-US"].validation.fields["test"] = attribute{Rules: map[string]string{"required": "test is required"}}
	suite.Equal("validation.fields.test", Get("en-US", "validation.fields.test"))
	suite.Equal("test is required", Get("en-US", "validation.fields.test.required"))
	suite.Equal("validation.fields.test.test", Get("en-US", "validation.fields.test.test"))

	languages["en-US"].validation.fields["test2"] = attribute{}
	suite.Equal("validation.fields.test2.required", Get("en-US", "validation.fields.test2.required"))

	suite.Equal("validation.fields", Get("en-US", "validation.fields"))
	suite.Equal("doesn't.exist", Get("not a language", "doesn't.exist"))
}

func (suite *LangTestSuite) TestDetectLanguage() {
	loadTestLang("fr-FR")
	loadTestLang("fr-FR") // Merge existing

	suite.Equal("en-US", DetectLanguage("en"))
	suite.Equal("en-US", DetectLanguage("en-US, fr"))
	suite.Equal("fr-FR", DetectLanguage("fr-FR"))
	suite.Equal("fr-FR", DetectLanguage("fr"))
	suite.Equal("en-US", DetectLanguage("fr, en-US"))
	suite.Equal("fr-FR", DetectLanguage("fr-FR, en-US"))
	suite.Equal("fr-FR", DetectLanguage("fr, en-US;q=0.9"))
	suite.Equal("en-US", DetectLanguage("en, fr-FR;q=0.9"))
	suite.Equal("en-US", DetectLanguage("*"))
	suite.Equal("en-US", DetectLanguage("notalang"))

	langs := GetAvailableLanguages()
	suite.Equal(2, len(langs))
	suite.Contains(langs, "en-US")
	suite.Contains(langs, "fr-FR")
}

func (suite *LangTestSuite) TestLoad() {
	suite.Panics(func() {
		Load("notalanguagedir", "notalanguagepath")
	})

	Load("en-US", "../resources/lang/en-US") // Is an override
	suite.Equal("rule override", languages["en-US"].validation.rules["required"])

	suite.Panics(func() {
		dest := map[string]string{}
		readLangFile("../resources/lang/invalid.json", &dest)
	})

	// Ensure default lang is not changed
	suite.Equal("The :field is required.", enUS.validation.rules["required"])

}

func (suite *LangTestSuite) TestMerge() {
	dst := language{
		lines: map[string]string{"line": "line 1"},
		validation: validationLines{
			rules: map[string]string{},
			fields: map[string]attribute{
				"test": {
					Name: "test field",
				},
			},
		},
	}
	src := language{
		lines: map[string]string{"other": "line 2"},
		validation: validationLines{
			rules: map[string]string{},
			fields: map[string]attribute{
				"email": {
					Name:  "email address",
					Rules: map[string]string{"required": "The email address is required"},
				},
				"test": {
					Name:  "test field override",
					Rules: map[string]string{"required": "The test field override is required"},
				},
			},
		},
	}

	mergeLang(dst, src)
	suite.Equal("line 1", dst.lines["line"])
	suite.Equal("line 2", dst.lines["other"])

	suite.Equal("email address", dst.validation.fields["email"].Name)
	suite.Equal("The email address is required", dst.validation.fields["email"].Rules["required"])

	suite.Equal("test field override", dst.validation.fields["test"].Name)
	suite.Equal("The test field override is required", dst.validation.fields["test"].Rules["required"])
}

func (suite *LangTestSuite) TearDownAllSuite() {
	languages = map[string]language{}
}

func TestLangTestSuite(t *testing.T) {
	suite.Run(t, new(LangTestSuite))
}
