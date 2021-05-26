package validation

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func validateString(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	_, ok := value.(string)
	return ok
}

func validateDigits(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		return getRegex(patternDigits).FindAllString(str, 1) == nil
	}
	return false
}

func validateAlpha(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	params := []string{patternAlpha}
	return validateRegex(field, value, params, form)
}

func validateAlphaDash(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	params := []string{patternAlphaDash}
	return validateRegex(field, value, params, form)
}

func validateAlphaNumeric(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	params := []string{patternAlphaNumeric}
	return validateRegex(field, value, params, form)
}

func validateEmail(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	params := []string{patternEmail}
	return validateRegex(field, value, params, form)
}

func validateStartsWith(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		for _, prefix := range parameters {
			if strings.HasPrefix(str, prefix) {
				return true
			}
		}
	}
	return false
}

func validateEndsWith(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		for _, prefix := range parameters {
			if strings.HasSuffix(str, prefix) {
				return true
			}
		}
	}
	return false
}

func validateIP(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		ip := net.ParseIP(str)
		if ip != nil {
			fieldName, _, parent, _ := GetFieldFromName(field, form)
			parent[fieldName] = ip
			return true
		}
	}

	return false
}

func validateIPv4(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if validateIP(field, value, parameters, form) {
		_, value, _, _ := GetFieldFromName(field, form)
		return value.(net.IP).To4() != nil
	}
	return false
}

func validateIPv6(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if validateIP(field, value, parameters, form) {
		_, value, _, _ := GetFieldFromName(field, form)
		return value.(net.IP).To4() == nil
	}
	return false
}

func validateJSON(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		var data interface{}
		err := json.Unmarshal([]byte(str), &data)
		if err == nil {
			fieldName, _, parent, _ := GetFieldFromName(field, form)
			parent[fieldName] = data
			return true
		}
	}
	return false
}

func validateRegex(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		return getRegex(parameters[0]).MatchString(str)
	}
	return false
}

func validateTimezone(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	tz, ok := value.(string)
	if ok && tz != "Local" {
		loc, err := time.LoadLocation(tz)
		if err == nil {
			fieldName, _, parent, _ := GetFieldFromName(field, form)
			parent[fieldName] = loc
			return true
		}
	}
	return false
}

func validateURL(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		url, err := url.ParseRequestURI(str)
		if err == nil {
			fieldName, _, parent, _ := GetFieldFromName(field, form)
			parent[fieldName] = url
			return true
		}
	}
	return false
}

func validateUUID(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	str, ok := value.(string)
	if ok {
		id, err := uuid.Parse(str)
		if err == nil {
			if len(parameters) == 1 {
				version, err := strconv.Atoi(parameters[0])
				if err == nil && id.Version() != uuid.Version(version) {
					return false
				}
			}
			fieldName, _, parent, _ := GetFieldFromName(field, form)
			parent[fieldName] = id
			return true
		}
	}
	return false
}
