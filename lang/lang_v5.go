package lang

import (
	"fmt"
	"os"
	"strings"

	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/httputil"
)

// Languages container for all loaded languages.
//
// This structure is not protected for concurrent usage. Therefore, don't load
// more languages when this instance is expected to receive reads.
type Languages struct {
	languages map[string]language
	Default   string
}

// TODO figure out a way to use embeds?

// New create a `Languages` with preloaded default language "en-US".
//
// The default language can be replaced by modifying the `Default` field
// in the returned struct.
func New() *Languages {
	l := &Languages{
		languages: make(map[string]language, 1),
		Default:   "en-US",
	}
	l.languages["en-US"] = enUS.clone()
	return l
}

// LoadAllAvailableLanguages loads every language directory
// in the "resources/lang" directory if it exists.
func (l *Languages) LoadAllAvailableLanguages() error {
	sep := string(os.PathSeparator)
	workingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	langDirectory := workingDir + sep + "resources" + sep + "lang" + sep
	if fsutil.IsDirectory(langDirectory) {
		files, err := os.ReadDir(langDirectory)
		if err != nil {
			return err
		}

		for _, f := range files {
			if f.IsDir() {
				if err := l.load(f.Name(), langDirectory+sep+f.Name()); err != nil {
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
//  en-UK
//    ├─ locale.json     (contains the normal language lines)
//    ├─ rules.json      (contains the validation messages)
//    └─ attributes.json (contains the attribute-specific validation messages)
//
// Each file is optional.
func (l *Languages) Load(language, path string) error {
	if fsutil.IsDirectory(path) {
		return l.load(language, path)
	}

	return fmt.Errorf("Failed loading language \"%s\", directory \"%s\" doesn't exist", language, path)
}

func (l *Languages) load(lang string, path string) error {
	langStruct := language{}
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

// Get a language line.
//
// For validation rules and attributes messages, use a dot-separated path:
// - "validation.rules.<rule_name>"
// - "validation.fields.<field_name>"
// - "validation.fields.<field_name>.<rule_name>"
// For normal lines, just use the name of the line. Note that if you have
// a line called "validation", it won't conflict with the dot-separated paths.
//
// If not found, returns the exact "line" attribute.
//
// The placeholders parameter is a variadic associative slice of placeholders and their
// replacement. In the following example, the placeholder ":username" will be replaced
// with the Name field in the user struct.
//
// 	lang.Get("en-US", "greetings", ":username", user.Name)
func (l *Languages) Get(lang string, line string, placeholders ...string) string {
	if !l.IsAvailable(lang) {
		return line
	}

	if strings.Count(line, ".") > 0 {
		path := strings.Split(line, ".")
		if path[0] == "validation" {
			switch path[1] {
			case "rules":
				if len(path) < 3 {
					return line
				}
				return convertEmptyLine(line, l.languages[lang].validation.rules[strings.Join(path[2:], ".")], placeholders)
			case "fields":
				len := len(path)
				if len < 3 {
					return line
				}
				attr := l.languages[lang].validation.fields[path[2]]
				if len == 4 {
					if attr.Rules == nil {
						return line
					}
					return convertEmptyLine(line, attr.Rules[path[3]], placeholders)
				} else if len == 3 {
					return convertEmptyLine(line, attr.Name, placeholders)
				} else {
					return line
				}
			default:
				return line
			}
		}
	}

	return convertEmptyLine(line, l.languages[lang].lines[line], placeholders)
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
//  /en/products
//  /fr/produits
//  ...
func (l *Languages) GetAvailableLanguages() []string {
	langs := []string{}
	for lang := range l.languages {
		langs = append(langs, lang)
	}
	return langs
}

// DetectLanguage detects the language to use based on the given lang string.
// The given lang string can use the HTTP "Accept-Language" header format.
//
// If "*" is provided, the default language will be used.
// If multiple languages are given, the first available language will be used,
// and if none are available, the default language will be used.
// If no variant is given (for example "en"), the first available variant will be used.
// For example, if "en-US" and "en-UK" are available and the request accepts "en",
// "en-US" will be used.
func (l *Languages) DetectLanguage(lang string) string {
	values := httputil.ParseMultiValuesHeader(lang)
	for _, lang := range values {
		if lang.Value == "*" { // Accept anything, so return default language
			break
		}
		if IsAvailable(lang.Value) {
			return lang.Value
		}
		for key := range l.languages {
			if strings.HasPrefix(key, lang.Value) {
				return key
			}
		}
	}

	return l.Default
}
