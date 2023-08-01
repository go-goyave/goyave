package lang

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/fsutil"
	"goyave.dev/goyave/v5/util/httputil"
)

// Languages container for all loaded languages.
//
// This structure is not protected for concurrent usage. Therefore, don't load
// more languages when this instance is expected to receive reads.
type Languages struct {
	languages map[string]*Language
	Default   string
}

// TODO figure out a way to use embeds?

// New create a `Languages` with preloaded default language "en-US".
//
// The default language can be replaced by modifying the `Default` field
// in the returned struct.
func New() *Languages {
	l := &Languages{
		languages: make(map[string]*Language, 1),
		Default:   enUS.name,
	}
	l.languages[enUS.name] = enUS.clone()
	return l
}

// LoadAllAvailableLanguages loads every language directory
// in the "resources/lang" directory if it exists.
func (l *Languages) LoadAllAvailableLanguages() error {
	workingDir, err := os.Getwd()
	if err != nil {
		return errors.New(err)
	}
	sep := string(os.PathSeparator)
	langDirectory := workingDir + sep + "resources" + sep + "lang" + sep
	return l.LoadDirectory(langDirectory)
}

// LoadDirectory loads every language directory
// in the given directory if it exists.
func (l *Languages) LoadDirectory(directory string) error {
	sep := string(os.PathSeparator)
	if fsutil.IsDirectory(directory) {
		files, err := os.ReadDir(directory)
		if err != nil {
			return errors.New(err)
		}

		for _, f := range files {
			if f.IsDir() {
				if err := l.load(f.Name(), directory+sep+f.Name()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Load a language directory.
//
// Directory structure of a language directory:
//
//	en-UK
//	  ├─ locale.json     (contains the normal language lines)
//	  ├─ rules.json      (contains the validation messages)
//	  └─ fields.json     (contains the field names)
//
// Each file is optional.
func (l *Languages) Load(language, path string) error {
	if fsutil.IsDirectory(path) {
		return l.load(language, path)
	}

	return errors.New(fmt.Errorf("failed loading language \"%s\", directory \"%s\" doesn't exist or is not readable", language, path))
}

func (l *Languages) load(lang string, path string) error {
	langStruct := &Language{
		name:  lang,
		lines: map[string]string{},
		validation: validationLines{
			rules:  map[string]string{},
			fields: map[string]string{},
		},
	}
	sep := string(os.PathSeparator)
	if err := readLangFile(path+sep+"locale.json", &langStruct.lines); err != nil {
		return err
	}
	if err := readLangFile(path+sep+"rules.json", &langStruct.validation.rules); err != nil {
		return err
	}
	if err := readLangFile(path+sep+"fields.json", &langStruct.validation.fields); err != nil {
		return err
	}

	if existingLang, exists := l.languages[lang]; exists {
		mergeLang(existingLang, langStruct)
	} else {
		l.languages[lang] = langStruct
	}
	return nil
}

// GetLanguage returns a language by its name.
// If the language is not available, returns a dummy language
// that will always return the entry name.
func (l *Languages) GetLanguage(lang string) *Language {
	if lang, ok := l.languages[lang]; ok {
		return lang
	}
	return &Language{
		name:  "dummy",
		lines: make(map[string]string, 0),
		validation: validationLines{
			rules:  make(map[string]string, 0),
			fields: make(map[string]string, 0),
		},
	}
}

// GetDefault is an alias for `l.GetLanguage(l.Default)`
func (l *Languages) GetDefault() *Language {
	return l.GetLanguage(l.Default)
}

// IsAvailable returns true if the language is available.
func (l *Languages) IsAvailable(lang string) bool {
	_, exists := l.languages[lang]
	return exists
}

// GetAvailableLanguages returns a slice of all loaded languages.
// This can be used to generate different routes for all languages
// supported by your applications.
//
//	/en/products
//	/fr/produits
//	...
func (l *Languages) GetAvailableLanguages() []string {
	return maps.Keys(l.languages)
}

// DetectLanguage detects the language to use based on the given lang string.
// The given lang string can use the HTTP "Accept-Language" header format.
//
// If "*" is provided, the default language will be used.
// If multiple languages are given, the first available language will be used,
// and if none are available, the default language will be used.
// If no variant is given (for example "en"), the first available variant will be used.
func (l *Languages) DetectLanguage(lang string) *Language {
	values := httputil.ParseMultiValuesHeader(lang)
	for _, lang := range values {
		if lang.Value == "*" { // Accept anything, so return default language
			break
		}
		if match, ok := l.languages[lang.Value]; ok {
			return match
		}
		for key, v := range l.languages {
			if strings.HasPrefix(key, lang.Value) {
				// TODO priority for languages? The "first available variant" is random because keys are not ordered.
				// Ordering alphabetically won't always produce the desired result (e.g. en-UK < en-US)
				// Can create a slice of language names (so the order will be the order in which the languages have been added)
				return v
			}
		}
	}

	return l.GetLanguage(l.Default)
}

// Get a language line.
//
// For validation rules messages and field names, use a dot-separated path:
//   - "validation.rules.<rule_name>"
//   - "validation.fields.<field_name>"
//
// For normal lines, just use the name of the line. Note that if you have
// a line called "validation", it won't conflict with the dot-separated paths.
//
// If not found, returns the exact "line" argument.
//
// The placeholders parameter is a variadic associative slice of placeholders and their
// replacement. In the following example, the placeholder ":username" will be replaced
// with the Name field in the user struct.
//
//	lang.Get("en-US", "greetings", ":username", user.Name)
func (l *Languages) Get(lang string, line string, placeholders ...string) string {
	language, exists := l.languages[lang]
	if !exists {
		return line
	}

	return language.Get(line, placeholders...)
}

func readLangFile(path string, dest any) (err error) {
	if !fsutil.FileExists(path) {
		return nil
	}

	langFile, _ := os.Open(path)
	defer func() {
		closeErr := langFile.Close()
		if err == nil && closeErr != nil {
			err = errors.New(closeErr)
		}
	}()

	err = json.NewDecoder(langFile).Decode(&dest)
	if err != nil {
		err = errors.New(err)
	}
	return
}

func mergeLang(dst *Language, src *Language) {
	mergeMap(dst.lines, src.lines)
	mergeMap(dst.validation.rules, src.validation.rules)
	mergeMap(dst.validation.fields, src.validation.fields)
}

func mergeMap(dst map[string]string, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
}
