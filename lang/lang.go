package lang

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/util/fsutil"
	"goyave.dev/goyave/v4/util/httputil"
)

type validationLines struct {
	// Default messages for rules
	rules map[string]string

	// Attribute-specific rules messages
	fields map[string]attribute
}

type attribute struct {
	// A custom message for when a rule doesn't pass with this attribute
	Rules map[string]string `json:"rules"`

	// The value with which the :field placeholder will be replaced
	Name string `json:"name"`
}

// language represents a full language
type language struct {
	lines      map[string]string
	validation validationLines
}

var languages map[string]language
var mutex = &sync.RWMutex{}

func (l *language) clone() language {
	cpy := language{
		lines: make(map[string]string, len(l.lines)),
		validation: validationLines{
			rules:  make(map[string]string, len(l.validation.rules)),
			fields: make(map[string]attribute, len(l.validation.fields)),
		},
	}

	mergeMap(cpy.lines, l.lines)
	mergeMap(cpy.validation.rules, l.validation.rules)

	for key, attr := range l.validation.fields {
		attrCpy := attribute{
			Name:  attr.Name,
			Rules: make(map[string]string, len(attr.Rules)),
		}
		mergeMap(attrCpy.Rules, attrCpy.Rules)
		cpy.validation.fields[key] = attrCpy
	}

	return cpy
}

// LoadDefault load the fallback language ("en-US").
// This function is intended for internal use only.
func LoadDefault() {
	mutex.Lock()
	defer mutex.Unlock()
	languages = make(map[string]language, 1)
	languages["en-US"] = enUS.clone()
}

// LoadAllAvailableLanguages loads every language directory
// in the "resources/lang" directory if it exists.
func LoadAllAvailableLanguages() {
	mutex.Lock()
	defer mutex.Unlock()
	sep := string(os.PathSeparator)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	langDirectory := workingDir + sep + "resources" + sep + "lang" + sep
	if fsutil.IsDirectory(langDirectory) {
		files, err := os.ReadDir(langDirectory)
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			if f.IsDir() {
				load(f.Name(), langDirectory+sep+f.Name())
			}
		}
	}
}

// Load a language directory.
//
// Directory structure of a language directory:
//
//	en-UK
//	  ├─ locale.json (contains the normal language lines)
//	  ├─ rules.json  (contains the validation messages)
//	  └─ fields.json (contains the attribute-specific validation messages)
//
// Each file is optional.
func Load(language, path string) {
	mutex.Lock()
	defer mutex.Unlock()
	if fsutil.IsDirectory(path) {
		load(language, path)
	} else {
		panic(fmt.Sprintf("Failed loading language \"%s\", directory \"%s\" doesn't exist", language, path))
	}
}

func load(lang string, path string) {
	langStruct := language{}
	sep := string(os.PathSeparator)
	readLangFile(path+sep+"locale.json", &langStruct.lines)
	readLangFile(path+sep+"rules.json", &langStruct.validation.rules)
	readLangFile(path+sep+"fields.json", &langStruct.validation.fields)

	if existingLang, exists := languages[lang]; exists {
		mergeLang(existingLang, langStruct)
	} else {
		languages[lang] = langStruct
	}
}

func readLangFile(path string, dest interface{}) {
	if fsutil.FileExists(path) {
		langFile, _ := os.Open(path)
		defer langFile.Close()

		errParse := json.NewDecoder(langFile).Decode(&dest)
		if errParse != nil {
			panic(errParse)
		}
	}
}

func mergeLang(dst language, src language) {
	mergeMap(dst.lines, src.lines)
	mergeMap(dst.validation.rules, src.validation.rules)

	for key, value := range src.validation.fields {
		if attr, exists := dst.validation.fields[key]; !exists {
			dst.validation.fields[key] = value
		} else {
			attr.Name = value.Name
			if attr.Rules == nil {
				attr.Rules = make(map[string]string)
			}
			mergeMap(attr.Rules, value.Rules)
			dst.validation.fields[key] = attr
		}
	}
}

func mergeMap(dst map[string]string, src map[string]string) {
	for key, value := range src {
		dst[key] = value
	}
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
//	lang.Get("en-US", "greetings", ":username", user.Name)
func Get(lang string, line string, placeholders ...string) string {
	if !IsAvailable(lang) {
		return line
	}

	mutex.RLock()
	defer mutex.RUnlock()
	if strings.Count(line, ".") > 0 {
		path := strings.Split(line, ".")
		if path[0] == "validation" {
			switch path[1] {
			case "rules":
				if len(path) < 3 {
					return line
				}
				return convertEmptyLine(line, languages[lang].validation.rules[strings.Join(path[2:], ".")], placeholders)
			case "fields":
				length := len(path)
				if length < 3 {
					return line
				}
				attr := languages[lang].validation.fields[path[2]]
				if length == 4 {
					if attr.Rules == nil {
						return line
					}
					return convertEmptyLine(line, attr.Rules[path[3]], placeholders)
				} else if length == 3 {
					return convertEmptyLine(line, attr.Name, placeholders)
				} else {
					return line
				}
			default:
				return line
			}
		}
	}

	return convertEmptyLine(line, languages[lang].lines[line], placeholders)
}

func processPlaceholders(message string, values []string) string {
	length := len(values) - 1
	for i := 0; i < length; i += 2 {
		message = strings.ReplaceAll(message, values[i], values[i+1])
	}
	return message
}

func convertEmptyLine(entry, line string, placeholders []string) string {
	if line == "" {
		return entry
	}
	return processPlaceholders(line, placeholders)
}

// IsAvailable returns true if the language is available.
func IsAvailable(lang string) bool {
	mutex.RLock()
	defer mutex.RUnlock()
	_, exists := languages[lang]
	return exists
}

// GetAvailableLanguages returns a slice of all loaded languages.
// This can be used to generate different routes for all languages
// supported by your applications.
//
//	/en/products
//	/fr/produits
//	...
func GetAvailableLanguages() []string {
	mutex.RLock()
	defer mutex.RUnlock()
	langs := []string{}
	for lang := range languages {
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
func DetectLanguage(lang string) string {
	values := httputil.ParseMultiValuesHeader(lang)
	for _, l := range values {
		if l.Value == "*" { // Accept anything, so return default language
			break
		}
		if IsAvailable(l.Value) {
			return l.Value
		}
		for key := range languages {
			if strings.HasPrefix(key, l.Value) {
				return key
			}
		}
	}

	return config.GetString("app.defaultLanguage")
}
