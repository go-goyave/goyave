package validation

import (
	"strconv"
	"strings"

	"github.com/System-Glitch/goyave/helper"
	"github.com/System-Glitch/goyave/helper/filesystem"
)

func validateFile(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.([]filesystem.File)
	return ok
}

func validateMIME(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("mime", parameters, 1)
	files, ok := value.([]filesystem.File)
	if ok {
		for _, file := range files {
			mime := file.MIMEType
			if i := strings.Index(mime, ";"); i != -1 { // Ignore MIME settings (example: "text/plain; charset=utf-8")
				mime = mime[:i]
			}
			if !helper.Contains(parameters, mime) {
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
	RequireParametersCount("extension", parameters, 1)
	files, ok := value.([]filesystem.File)
	if ok {
		for _, file := range files {
			if i := strings.LastIndex(file.Header.Filename, "."); i != -1 {
				if !helper.Contains(parameters, file.Header.Filename[i+1:]) {
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

func validateCount(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("count", parameters, 1)
	files, ok := value.([]filesystem.File)
	size, err := strconv.Atoi(parameters[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) == size
	}

	return false
}

func validateCountMin(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("count_min", parameters, 1)
	files, ok := value.([]filesystem.File)
	size, err := strconv.Atoi(parameters[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) >= size
	}

	return false
}

func validateCountMax(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("count_max", parameters, 1)
	files, ok := value.([]filesystem.File)
	size, err := strconv.Atoi(parameters[0])
	if err != nil {
		panic(err)
	}

	if ok {
		return len(files) <= size
	}

	return false
}

func validateCountBetween(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("count_between", parameters, 1)
	files, ok := value.([]filesystem.File)
	min, errMin := strconv.Atoi(parameters[0])
	max, errMax := strconv.Atoi(parameters[1])
	if errMin != nil {
		panic(errMin)
	}
	if errMax != nil {
		panic(errMax)
	}

	if ok {
		length := len(files)
		return length >= min && length <= max
	}

	return false
}
