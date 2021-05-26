package validation

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/Code-Hex/uniseg"
	"goyave.dev/goyave/v3/helper"
	"goyave.dev/goyave/v3/helper/filesystem"
)

func validateRequired(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, val, _, ok := GetFieldFromName(field, form)
	if ok {
		if str, okStr := val.(string); okStr && str == "" {
			return false
		}
	}
	return ok
}

func validateMin(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	min, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch GetFieldType(value) {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		return floatValue >= min
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) >= int(min)
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
	max, err := strconv.ParseFloat(parameters[0], 64)
	if err != nil {
		panic(err)
	}
	switch GetFieldType(value) {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		return floatValue <= max
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) <= int(max)
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
	min, errMin := strconv.ParseFloat(parameters[0], 64)
	max, errMax := strconv.ParseFloat(parameters[1], 64)
	if errMin != nil {
		panic(errMin)
	}
	if errMax != nil {
		panic(errMax)
	}

	switch GetFieldType(value) {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		return floatValue >= min && floatValue <= max
	case "string":
		length := uniseg.GraphemeClusterCount(value.(string))
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
	valueType := GetFieldType(value)

	_, compared, _, exists := GetFieldFromName(parameters[0], form)
	if !exists || valueType != GetFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		comparedFloatValue, _ := helper.ToFloat64(compared)
		return floatValue > comparedFloatValue
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) > uniseg.GraphemeClusterCount(compared.(string))
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

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateGreaterThanEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	valueType := GetFieldType(value)

	_, compared, _, exists := GetFieldFromName(parameters[0], form)
	if !exists || valueType != GetFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		comparedFloatValue, _ := helper.ToFloat64(compared)
		return floatValue >= comparedFloatValue
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) >= uniseg.GraphemeClusterCount(compared.(string))
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

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateLowerThan(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	valueType := GetFieldType(value)

	_, compared, _, exists := GetFieldFromName(parameters[0], form)
	if !exists || valueType != GetFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		comparedFloatValue, _ := helper.ToFloat64(compared)
		return floatValue < comparedFloatValue
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) < uniseg.GraphemeClusterCount(compared.(string))
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

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateLowerThanEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	valueType := GetFieldType(value)

	_, compared, _, exists := GetFieldFromName(parameters[0], form)
	if !exists || valueType != GetFieldType(compared) {
		return false // Can't compare two different types or missing field
	}

	switch valueType {
	case "numeric":
		floatValue, _ := helper.ToFloat64(value)
		comparedFloatValue, _ := helper.ToFloat64(compared)
		return floatValue <= comparedFloatValue
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) <= uniseg.GraphemeClusterCount(compared.(string))
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

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateBool(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	fieldName, _, parent, _ := GetFieldFromName(field, form)
	switch {
	case kind == "bool":
		return true
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr":
		v, _ := helper.ToFloat64(value)
		if v == 1 {
			parent[fieldName] = true
			return true
		} else if v == 0 {
			parent[fieldName] = false
			return true
		}
	case kind == "string":
		v, _ := value.(string)
		switch v {
		case "1", "on", "true", "yes":
			parent[fieldName] = true
			return true
		case "0", "off", "false", "no":
			parent[fieldName] = false
			return true
		}
	}
	return false
}

func validateSame(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, other, _, exists := GetFieldFromName(parameters[0], form)
	if exists {
		valueType := GetFieldType(value)
		otherType := GetFieldType(other)
		if valueType == otherType {
			switch valueType {
			case "numeric":
				f1, _ := helper.ToFloat64(value)
				f2, _ := helper.ToFloat64(other)
				return f1 == f2
			case "string":
				s1, _ := value.(string)
				s2, _ := other.(string)
				return s1 == s2
			case "array":
				return helper.SliceEqual(value, other)
			case "object":
				return reflect.DeepEqual(value, other)
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
	params := []string{field + "_confirmation"}
	return validateSame(field, value, params, form)
}

func validateSize(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	size, err := strconv.Atoi(parameters[0])
	if err != nil {
		panic(err)
	}

	switch GetFieldType(value) {
	case "numeric":
		floatVal, _ := helper.ToFloat64(value)
		return floatVal == float64(size)
	case "string":
		return uniseg.GraphemeClusterCount(value.(string)) == size
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

	return true // Pass if field type cannot be checked (bool, dates, ...)
}

func validateObject(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.(map[string]interface{})
	return ok
}
