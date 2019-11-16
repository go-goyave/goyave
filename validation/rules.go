package validation

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/System-Glitch/goyave/helpers"
	"github.com/System-Glitch/goyave/helpers/filesystem"
)

func validateRequired(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	val, ok := form[field]
	if ok {
		if str, okStr := val.(string); okStr && str == "" {
			return false
		}
	}
	return ok
}

func validateMin(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("min", parameters, 1)
	min, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch getFieldType(value) {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		return floatValue >= min
	case "string":
		return len(value.(string)) >= int(min)
	case "array":
		list := reflect.ValueOf(value)
		return list.Len() >= int(min)
	case "file":
		files, _ := value.([]filesystem.File)
		for _, file := range files {
			if file.Header.Size < int64(min)*1024 {
				return false
			}
		}
		return true
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateMax(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("max", parameters, 1)
	max, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch getFieldType(value) {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		return floatValue <= max
	case "string":
		return len(value.(string)) <= int(max)
	case "array":
		list := reflect.ValueOf(value)
		return list.Len() <= int(max)
	case "file":
		files, _ := value.([]filesystem.File)
		for _, file := range files {
			if file.Header.Size > int64(max)*1024 {
				return false
			}
		}
		return true
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateBetween(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("between", parameters, 2)
	min, errMin := strconv.ParseFloat(parameters[0], 64)
	max, errMax := strconv.ParseFloat(parameters[1], 64)
	if errMin != nil {
		panic(errMin)
	}
	if errMax != nil {
		panic(errMax)
	}

	switch getFieldType(value) {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		return floatValue >= min && floatValue <= max
	case "string":
		length := len(value.(string))
		return length >= int(min) && length <= int(max)
	case "array":
		list := reflect.ValueOf(value)
		length := list.Len()
		return length >= int(min) && length <= int(max)
	case "file":
		files, _ := value.([]filesystem.File)
		for _, file := range files {
			minSize := int64(min) * 1024
			maxSize := int64(max) * 1024
			if file.Header.Size < minSize || file.Header.Size > maxSize {
				return false
			}
		}
		return true
	}

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateGreaterThan(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("greater_than", parameters, 1)
	valueType := getFieldType(value)

	compared, exists := form[parameters[0]]
	if !exists || valueType != getFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		comparedFloatValue, _ := helpers.ToFloat64(compared)
		return floatValue > comparedFloatValue
	case "string":
		return len(value.(string)) > len(compared.(string))
	case "array":
		return reflect.ValueOf(value).Len() > reflect.ValueOf(compared).Len()
	case "file":
		files, _ := value.([]filesystem.File)
		comparedFiles, _ := compared.([]filesystem.File)
		for _, file := range files {
			for _, comparedFile := range comparedFiles {
				if file.Header.Size <= comparedFile.Header.Size {
					return false
				}
			}
		}
		return true
	}

	return false
}

func validateGreaterThanEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("greater_than_equal", parameters, 1)
	valueType := getFieldType(value)

	compared, exists := form[parameters[0]]
	if !exists || valueType != getFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		comparedFloatValue, _ := helpers.ToFloat64(compared)
		return floatValue >= comparedFloatValue
	case "string":
		return len(value.(string)) >= len(compared.(string))
	case "array":
		return reflect.ValueOf(value).Len() >= reflect.ValueOf(compared).Len()
	case "file":
		files, _ := value.([]filesystem.File)
		comparedFiles, _ := compared.([]filesystem.File)
		for _, file := range files {
			for _, comparedFile := range comparedFiles {
				if file.Header.Size < comparedFile.Header.Size {
					return false
				}
			}
		}
		return true
	}

	return false
}

func validateLowerThan(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("lower_than", parameters, 1)
	valueType := getFieldType(value)

	compared, exists := form[parameters[0]]
	if !exists || valueType != getFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		comparedFloatValue, _ := helpers.ToFloat64(compared)
		return floatValue < comparedFloatValue
	case "string":
		return len(value.(string)) < len(compared.(string))
	case "array":
		return reflect.ValueOf(value).Len() < reflect.ValueOf(compared).Len()
	case "file":
		files, _ := value.([]filesystem.File)
		comparedFiles, _ := compared.([]filesystem.File)
		for _, file := range files {
			for _, comparedFile := range comparedFiles {
				if file.Header.Size >= comparedFile.Header.Size {
					return false
				}
			}
		}
		return true
	}

	return false
}

func validateLowerThanEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("lower_than_equal", parameters, 1)
	valueType := getFieldType(value)

	compared, exists := form[parameters[0]]
	if !exists || valueType != getFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helpers.ToFloat64(value)
		comparedFloatValue, _ := helpers.ToFloat64(compared)
		return floatValue <= comparedFloatValue
	case "string":
		return len(value.(string)) <= len(compared.(string))
	case "array":
		return reflect.ValueOf(value).Len() <= reflect.ValueOf(compared).Len()
	case "file":
		files, _ := value.([]filesystem.File)
		comparedFiles, _ := compared.([]filesystem.File)
		for _, file := range files {
			for _, comparedFile := range comparedFiles {
				if file.Header.Size > comparedFile.Header.Size {
					return false
				}
			}
		}
		return true
	}

	return false
}

func validateBool(field string, value interface{}, parameters []string, form map[string]interface{}) bool { // TODO document accepted values and convert
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	switch {
	case kind == "bool":
		return true
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		v, _ := helpers.ToFloat64(value)
		if v == 1 {
			form[field] = true
			return true
		} else if v == 0 {
			form[field] = false
			return true
		}
	case kind == "string":
		v, _ := value.(string)
		switch v {
		case "on", "true", "yes":
			form[field] = true
			return true
		case "off", "false", "no":
			form[field] = false
			return true
		}
	}
	return false
}

func validateSame(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("same", parameters, 1)
	other, exists := form[parameters[0]]
	if exists {
		valueType := getFieldType(value)
		otherType := getFieldType(other)
		if valueType == otherType {
			switch valueType {
			case "numeric":
				f1, _ := helpers.ToFloat64(value)
				f2, _ := helpers.ToFloat64(other)
				return f1 == f2
			case "string":
				s1, _ := value.(string)
				s2, _ := other.(string)
				return s1 == s2
			case "array":
				return helpers.SliceEqual(value, other)
			}
			// Don't check files
		}
	}
	return false
}

func validateDifferent(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	return !validateSame(field, value, parameters, form)
}

func validateConfirmed(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	parameters = []string{field + "_confirmation"}
	return validateSame(field, value, parameters, form)
}

func validateSize(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	RequireParametersCount("size", parameters, 1)
	size, err := strconv.Atoi(parameters[0])
	if err != nil {
		panic(err)
	}

	switch getFieldType(value) {
	case "numeric":
		floatVal, _ := helpers.ToFloat64(value)
		return floatVal == float64(size)
	case "string":
		return len(value.(string)) == size
	case "array":
		list := reflect.ValueOf(value)
		return list.Len() == size
	case "file":
		files, _ := value.([]filesystem.File)
		for _, file := range files {
			if int64(math.Round(float64(file.Header.Size)/1024.0)) != int64(size) {
				return false
			}
		}
		return true
	}
	return false
}
