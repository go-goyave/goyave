package semconv

import (
	"net/http"
	"strings"
)

// parseForwardedHeader returns the parameter of the first (client-most)
// element of the RFC 7239 Forwarded header, or "" if there is none.
//
// Only the first element is considered: later elements are added by
// proxies further from the client and would carry internal hostnames.
func parseForwardedHeader(h http.Header, parameterName string) string { // TODO move this to httputil?
	for _, line := range h.Values("Forwarded") {
		for _, elem := range splitUnquoted(line, ',') {
			if strings.TrimSpace(elem) == "" {
				continue
			}
			for _, pair := range splitUnquoted(elem, ';') {
				k, v, ok := strings.Cut(pair, "=")
				if ok && strings.EqualFold(strings.TrimSpace(k), parameterName) {
					return unquote(strings.TrimSpace(v))
				}
			}
			return "" // first element has no such parameter: don't use later hops
		}
	}
	return ""
}

// splitUnquoted splits s on sep, ignoring separators inside quoted strings.
func splitUnquoted(s string, sep byte) []string {
	var parts []string
	inQuote, escaped, start := false, false, 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case escaped:
			escaped = false
		case inQuote && c == '\\':
			escaped = true
		case c == '"':
			inQuote = !inQuote
		case c == sep && !inQuote:
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	return append(parts, s[start:])
}

// unquote removes surrounding quotes and backslash escapes (RFC 9110 quoted-string).
func unquote(v string) string {
	if len(v) < 2 || v[0] != '"' || v[len(v)-1] != '"' {
		return v
	}
	var b strings.Builder
	for i := 1; i < len(v)-1; i++ {
		if v[i] == '\\' && i+1 < len(v)-1 {
			i++
		}
		b.WriteByte(v[i])
	}
	return b.String()
}
