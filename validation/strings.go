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

func validateString(ctx *Context) bool {
	_, ok := ctx.Value.(string)
	return ok
}

func validateDigits(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		return getRegex(patternDigits).FindAllString(str, 1) == nil
	}
	return false
}

func validateAlpha(ctx *Context) bool {
	ctx.Rule.Params = []string{patternAlpha}
	return validateRegex(ctx)
}

func validateAlphaDash(ctx *Context) bool {
	ctx.Rule.Params = []string{patternAlphaDash}
	return validateRegex(ctx)
}

func validateAlphaNumeric(ctx *Context) bool {
	ctx.Rule.Params = []string{patternAlphaNumeric}
	return validateRegex(ctx)
}

func validateEmail(ctx *Context) bool {
	ctx.Rule.Params = []string{patternEmail}
	return validateRegex(ctx)
}

func validateStartsWith(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		for _, prefix := range ctx.Rule.Params {
			if strings.HasPrefix(str, prefix) {
				return true
			}
		}
	}
	return false
}

func validateEndsWith(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		for _, prefix := range ctx.Rule.Params {
			if strings.HasSuffix(str, prefix) {
				return true
			}
		}
	}
	return false
}

func validateIP(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		ip := net.ParseIP(str)
		if ip != nil {
			ctx.Value = ip
			return true
		}
	}

	return false
}

func validateIPv4(ctx *Context) bool {
	if validateIP(ctx) {
		return ctx.Value.(net.IP).To4() != nil
	}
	return false
}

func validateIPv6(ctx *Context) bool {
	if validateIP(ctx) {
		return ctx.Value.(net.IP).To4() == nil
	}
	return false
}

func validateJSON(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		var data interface{}
		err := json.Unmarshal([]byte(str), &data)
		if err == nil {
			ctx.Value = data
			return true
		}
	}
	return false
}

func validateRegex(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		return getRegex(ctx.Rule.Params[0]).MatchString(str)
	}
	return false
}

func validateTimezone(ctx *Context) bool {
	tz, ok := ctx.Value.(string)
	if ok && tz != "Local" {
		loc, err := time.LoadLocation(tz)
		if err == nil {
			ctx.Value = loc
			return true
		}
	}
	return false
}

func validateURL(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		url, err := url.ParseRequestURI(str)
		if err == nil {
			ctx.Value = url
			return true
		}
	}
	return false
}

func validateUUID(ctx *Context) bool {
	str, ok := ctx.Value.(string)
	if ok {
		id, err := uuid.Parse(str)
		if err == nil {
			if len(ctx.Rule.Params) == 1 {
				version, err := strconv.Atoi(ctx.Rule.Params[0])
				if err == nil && id.Version() != uuid.Version(version) {
					return false
				}
			}
			ctx.Value = id
			return true
		}
	}
	return false
}
