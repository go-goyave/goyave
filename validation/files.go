package validation

import (
	"strings"

	"github.com/System-Glitch/goyave/helpers"
	"github.com/System-Glitch/goyave/helpers/filesystem"
)

func validateFile(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.([]filesystem.File)
	return ok
}

func validateMIME(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	requireParametersCount("mime", parameters, 1)
	files, ok := value.([]filesystem.File)
	if ok {
		for _, file := range files {
			mime := file.MIMEType
			if i := strings.Index(mime, ";"); i != -1 { // Ignore MIME settings (example: "text/plain; charset=utf-8")
				mime = mime[:i]
			}
			if !helpers.Contains(parameters, mime) {
				return false
			}
		}
		return true
	}
	return false
}

func validateImage(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	parameters = []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/svg+xml", "image/webp"}
	return validateMIME(field, value, parameters, form)
}

func validateExtension(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	requireParametersCount("extension", parameters, 1)
	files, ok := value.([]filesystem.File)
	if ok {
		for _, file := range files {
			if i := strings.LastIndex(file.Header.Filename, "."); i != -1 {
				if !helpers.Contains(parameters, file.Header.Filename[i+1:]) {
					return false
				}
			} else {
				return false
			}
		}
		return true
	}
	return false
}
