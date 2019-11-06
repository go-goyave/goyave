package lang

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/System-Glitch/goyave/config"

	"github.com/System-Glitch/goyave/helpers/filesystem"
)

type validationLines struct {
	// Default messages for rules
	rules map[string]string

	// Attribute-specific rules messages
	fields map[string]attribute
}

type attribute struct {
	// The value with which the :field placeholder will be replaced
	Name string `json:"name"`

	// A custom message for when a rule doesn't pass with this attribute
	Rules map[string]string `json:"rules"`
}

// language represents a full language
type language struct {
	lines      map[string]string
	validation validationLines
}

var languages map[string]language = map[string]language{}

// LoadDefault load the fallback language ("en-US") and, if needed,
// the default language provided in the config.
// This function is intended for internal use only.
func LoadDefault() {
	var filename string
	var ok bool
	func() {
		_, f, _, o := runtime.Caller(1)
		filename = f
		ok = o
	}()

	if !ok {
		log.Panicf("Runtime caller error")
	}

	sep := string(os.PathSeparator)
	load("en-US", path.Dir(filename)+sep+".."+sep+"resources"+sep+"lang"+sep+"en-US")
}

// LoadAllAvailableLanguages loads every language directory
// in the "resources/lang" directory if it exists.
func LoadAllAvailableLanguages() {
	sep := string(os.PathSeparator)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	langDirectory := workingDir + sep + "resources" + sep + "lang" + sep
	if filesystem.IsDirectory(langDirectory) {
		files, err := ioutil.ReadDir(langDirectory)
		if err != nil {
			panic(err)
		}

		for _, f := range files {
			if f.IsDir() {
				Load(f.Name(), langDirectory)
			}
		}
	}
}

// Load a language directory.
//
// Directory structure of a language directory:
//  en-UK
//    |- locale.json     (contains the normal language lines)
//    |- rules.json      (contains the validation messages)
//    |- attributes.json (contains the attribute-specific validation messages)
//
// Each file is optional.
func Load(language, path string) {
	if filesystem.IsDirectory(path) {
		load(language, path)
	} else {
		log.Panicf("Failed loading language \"%s\", directory \"%s\" doesn't exist", language, path)
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
	if filesystem.FileExists(path) {
		langFile, err := os.Open(path)
		defer langFile.Close()

		if err == nil {
			errParse := json.NewDecoder(langFile).Decode(&dest)
			if errParse != nil {
				panic(errParse)
			}
		} else {
			panic(err)
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

// Get a language entry.
//
// For validation rules and attributes messages, use a dot-separated path:
// - "validation.rules.<rule_name>"
// - "validation.attributes.<attribute_name>.<rule_name>"
// For normal lines, just use the name of the line. Note that if you have
// a line called "validation", it won't conflict with the dot-separated paths.
//
// If not found, returns the exact "line" attribute.
func Get(lang string, line string) string {
	if !IsAvailable(lang) {
		return line
	}
	if strings.Count(line, ".") > 0 {
		path := strings.Split(line, ".")
		if path[0] != "validation" {
			return line
		}

		switch path[1] {
		case "rules":
			if len(path) < 3 {
				return line
			}
			return convertEmptyLine(line, languages[lang].validation.rules[strings.Join(path[2:], ".")])
		case "fields":
			len := len(path)
			if len < 3 {
				return line
			}
			attr := languages[lang].validation.fields[path[2]]
			if len == 4 {
				if attr.Rules == nil {
					return line
				}
				return convertEmptyLine(line, attr.Rules[path[3]])
			} else if len == 3 {
				return convertEmptyLine(line, attr.Name)
			} else {
				return line
			}
		default:
			return line
		}
	}

	return convertEmptyLine(line, languages[lang].lines[line])
}

func convertEmptyLine(entry, line string) string {
	if line == "" {
		return entry
	}
	return line
}

// IsAvailable returns true if the language is available.
func IsAvailable(lang string) bool {
	_, exists := languages[lang]
	return exists
}

// DetectLanguage detects the language to use based on the given lang string.
//
// If "*" is provided, the default language will be used.
// If multiple languages are given, the first available language will be used,
// and if none are available, the default language will be used.
// If no variant is given (for example "en"), the first available variant will be used.
// For example, if "en-US" and "en-UK" are available and the request accepts "en",
// "en-US" will be used.
func DetectLanguage(lang string) string {
	switch {
	case lang == "*":
		return config.GetString("defaultLanguage")
	case strings.Count(lang, ",") != 0: // Multiple languages
		if i := strings.Index(lang, ";"); i != -1 {
			lang = lang[:i]
		}
		for _, l := range strings.Split(lang, ",") {
			for key := range languages {
				if strings.HasPrefix(key, l) {
					return key
				}
			}
		}
	case strings.Count(lang, "-") == 0: // No variant
		for key := range languages {
			if strings.HasPrefix(key, lang) {
				return key
			}
		}
	}

	if IsAvailable(lang) {
		return lang
	}
	return config.GetString("defaultLanguage")
}
