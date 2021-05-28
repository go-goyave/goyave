package validation

import (
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateString(t *testing.T) {
	assert.True(t, validateString("field", "string", []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", 2, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", 2.5, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", []byte{}, []string{}, map[string]interface{}{}))
	assert.False(t, validateString("field", []string{}, []string{}, map[string]interface{}{}))
}

func TestValidateDigits(t *testing.T) {
	assert.True(t, validateDigits("field", "123", []string{}, map[string]interface{}{}))
	assert.True(t, validateDigits("field", "0123456789", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "2.3", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "-123", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "abcd", []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", "/*-&é\"'(-è_ç", []string{}, map[string]interface{}{}))

	// Not string
	assert.False(t, validateDigits("field", 1, []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", 1.2, []string{}, map[string]interface{}{}))
	assert.False(t, validateDigits("field", true, []string{}, map[string]interface{}{}))
}

func TestValidateRegex(t *testing.T) {
	assert.True(t, validateRegex("field", "sghtyhg", []string{"t"}, map[string]interface{}{}))
	assert.True(t, validateRegex("field", "sghtyhg", []string{"[^\\s]"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", "sgh tyhg", []string{"^[^\\s]+$"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", "48s9", []string{"^[^0-9]+$"}, map[string]interface{}{}))
	assert.True(t, validateRegex("field", "489", []string{"^[0-9]+$"}, map[string]interface{}{}))
	assert.False(t, validateRegex("field", 489, []string{"^[^0-9]+$"}, map[string]interface{}{}))

	assert.Panics(t, func() { validateRegex("field", "", []string{"doesn't compile \\"}, map[string]interface{}{}) })

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "regex"},
			},
		}
		field.check()
	})
}

func TestValidateEmail(t *testing.T) {
	assert.True(t, validateEmail("field", "simple@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "very.common@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "disposable.style.email.with+symbol@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "other.email-with-hyphen@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "fully-qualified-domain@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "user.name+tag+sorting@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "x@example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "example-indeed@strange-example.com", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "admin@mailserver1", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "example@s.example", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "\" \"@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "\"john..doe\"@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "mailhost!username@example.org", []string{}, map[string]interface{}{}))
	assert.True(t, validateEmail("field", "user%example.com@example.org", []string{}, map[string]interface{}{}))
	assert.False(t, validateEmail("field", "Abc.example.com", []string{}, map[string]interface{}{}))
	assert.False(t, validateEmail("field", "1234567890123456789012345678901234567890123456789012345678901234+x@example.com", []string{}, map[string]interface{}{}))
}

func TestValidateAlpha(t *testing.T) {
	assert.True(t, validateAlpha("field", "helloworld", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlpha("field", "éèçàû", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlpha("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateAlphaDash(t *testing.T) {
	assert.True(t, validateAlphaDash("field", "helloworld", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "éèçàû_-", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "hello-world", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaDash("field", "hello-world_2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaDash("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateAlphaNumeric(t *testing.T) {
	assert.True(t, validateAlphaNumeric("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaNumeric("field", "éèçàû2", []string{}, map[string]interface{}{}))
	assert.True(t, validateAlphaNumeric("field", "helloworld2", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", "hello world", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{}))
	assert.False(t, validateAlphaNumeric("field", 2, []string{}, map[string]interface{}{}))
}

func TestValidateStartsWith(t *testing.T) {
	assert.True(t, validateStartsWith("field", "hello world", []string{"hello"}, map[string]interface{}{}))
	assert.True(t, validateStartsWith("field", "hi", []string{"hello", "hi", "hey"}, map[string]interface{}{}))
	assert.False(t, validateStartsWith("field", "sup'!", []string{"hello", "hi", "hey"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "starts_with"},
			},
		}
		field.check()
	})
}

func TestValidateEndsWith(t *testing.T) {
	assert.True(t, validateEndsWith("field", "hello world", []string{"world"}, map[string]interface{}{}))
	assert.True(t, validateEndsWith("field", "oh hi mark", []string{"ross", "mark", "bruce"}, map[string]interface{}{}))
	assert.False(t, validateEndsWith("field", "sup' bro!", []string{"ross", "mark", "bruce"}, map[string]interface{}{}))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "ends_with"},
			},
		}
		field.check()
	})
}

func TestValidateTimezone(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateTimezone("field", "UTC", []string{}, data))
	assert.True(t, validateTimezone("field", "Europe/Paris", []string{}, data))
	assert.True(t, validateTimezone("field", "America/St_Thomas", []string{}, data))
	assert.True(t, validateTimezone("field", "GMT", []string{}, data))
	assert.False(t, validateTimezone("field", "GMT+2", []string{}, data))
	assert.False(t, validateTimezone("field", "UTC+2", []string{}, data))
	assert.False(t, validateTimezone("field", "here", []string{}, data))
	assert.False(t, validateTimezone("field", 1, []string{}, data))
	assert.False(t, validateTimezone("field", 1.5, []string{}, data))
	assert.False(t, validateTimezone("field", true, []string{}, data))
	assert.False(t, validateTimezone("field", []string{"UTC"}, []string{}, data))
}

func TestValidateTimezoneConvert(t *testing.T) {
	form := map[string]interface{}{"field": "UTC"}
	assert.True(t, validateTimezone("field", form["field"], []string{}, form))

	_, ok := form["field"].(*time.Location)
	assert.True(t, ok)
}

func TestValidateTimezoneConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"timezone": "UTC",
		},
	}

	set := RuleSet{
		"object":          {"required", "object"},
		"object.timezone": {"required", "timezone"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["timezone"].(*time.Location)
	assert.True(t, ok)
}

func TestValidateIP(t *testing.T) {
	data := map[string]interface{}{
		"field": "127.0.0.1",
	}
	assert.True(t, validateIP("field", "127.0.0.1", []string{}, data))
	assert.True(t, validateIP("field", "192.168.0.1", []string{}, data))
	assert.True(t, validateIP("field", "88.88.88.88", []string{}, data))
	assert.True(t, validateIP("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data))
	assert.True(t, validateIP("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data))
	assert.True(t, validateIP("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data))
	assert.True(t, validateIP("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data))
	assert.True(t, validateIP("field", "::1", []string{}, data))

	assert.False(t, validateIP("field", "1", []string{}, data))
	assert.False(t, validateIP("field", 1, []string{}, data))
	assert.False(t, validateIP("field", 1.2, []string{}, data))
	assert.False(t, validateIP("field", true, []string{}, data))
	assert.False(t, validateIP("field", []byte{}, []string{}, data))
}

func TestValidateIPConvert(t *testing.T) {
	form := map[string]interface{}{"field": "127.0.0.1"}
	assert.True(t, validateIP("field", form["field"], []string{}, form))

	_, ok := form["field"].(net.IP)
	assert.True(t, ok)
}

func TestValidateIPConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "127.0.0.1",
		},
	}

	set := RuleSet{
		"object":    {"required", "object"},
		"object.ip": {"required", "ip"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["ip"].(net.IP)
	assert.True(t, ok)
}

func TestValidateIPv4(t *testing.T) {
	data := map[string]interface{}{
		"field": "127.0.0.1",
	}
	assert.True(t, validateIPv4("field", "127.0.0.1", []string{}, data))
	assert.True(t, validateIPv4("field", "192.168.0.1", []string{}, data))
	assert.True(t, validateIPv4("field", "88.88.88.88", []string{}, data))
	assert.False(t, validateIPv4("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data))
	assert.False(t, validateIPv4("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data))
	assert.False(t, validateIPv4("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data))
	assert.False(t, validateIPv4("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data))
	assert.False(t, validateIPv4("field", "::1", []string{}, data))
	assert.False(t, validateIPv4("field", 1, []string{}, data))
	assert.False(t, validateIPv4("field", 1.2, []string{}, data))
	assert.False(t, validateIPv4("field", true, []string{}, data))
	assert.False(t, validateIPv4("field", []byte{}, []string{}, data))
}

func TestValidateIPv4ConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "127.0.0.1",
		},
	}

	set := RuleSet{
		"object":    {"required", "object"},
		"object.ip": {"required", "ipv4"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["ip"].(net.IP)
	assert.True(t, ok)
}

func TestValidateIPv6(t *testing.T) {
	data := map[string]interface{}{
		"field": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	}
	assert.False(t, validateIPv6("field", "127.0.0.1", []string{}, data))
	assert.False(t, validateIPv6("field", "192.168.0.1", []string{}, data))
	assert.False(t, validateIPv6("field", "88.88.88.88", []string{}, data))
	assert.True(t, validateIPv6("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data))
	assert.True(t, validateIPv6("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data))
	assert.True(t, validateIPv6("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data))
	assert.True(t, validateIPv6("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data))
	assert.True(t, validateIPv6("field", "::1", []string{}, data))
	assert.False(t, validateIPv6("field", 1, []string{}, data))
	assert.False(t, validateIPv6("field", 1.2, []string{}, data))
	assert.False(t, validateIPv6("field", true, []string{}, data))
	assert.False(t, validateIPv6("field", []byte{}, []string{}, data))
}

func TestValidateIPv6ConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		},
	}

	set := RuleSet{
		"object":    {"required", "object"},
		"object.ip": {"required", "ipv6"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["ip"].(net.IP)
	assert.True(t, ok)
}

func TestValidateJSON(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateJSON("field", "2", []string{}, data))
	assert.True(t, validateJSON("field", "2.5", []string{}, data))
	assert.True(t, validateJSON("field", "\"str\"", []string{}, data))
	assert.True(t, validateJSON("field", "[\"str\",\"array\"]", []string{}, data))
	assert.True(t, validateJSON("field", "{\"str\":\"object\"}", []string{}, data))
	assert.True(t, validateJSON("field", "{\"str\":[\"object\",\"array\"]}", []string{}, data))

	assert.False(t, validateJSON("field", "{str:[\"object\",\"array\"]}", []string{}, data))
	assert.False(t, validateJSON("field", "", []string{}, data))
	assert.False(t, validateJSON("field", "\"d", []string{}, data))
	assert.False(t, validateJSON("field", 1, []string{}, data))
	assert.False(t, validateJSON("field", 1.2, []string{}, data))
	assert.False(t, validateJSON("field", map[string]string{}, []string{}, data))
}

func TestValidateJSONConvert(t *testing.T) {
	form := map[string]interface{}{"field": "2"}
	assert.True(t, validateJSON("field", form["field"], []string{}, form))
	_, ok := form["field"].(float64)
	assert.True(t, ok)

	form = map[string]interface{}{"field": "\"str\""}
	assert.True(t, validateJSON("field", form["field"], []string{}, form))
	_, ok = form["field"].(string)
	assert.True(t, ok)

	form = map[string]interface{}{"field": "[\"str\",\"array\"]"}
	assert.True(t, validateJSON("field", form["field"], []string{}, form))
	_, ok = form["field"].([]interface{})
	assert.True(t, ok)

	form = map[string]interface{}{"field": "{\"str\":\"object\"}"}
	assert.True(t, validateJSON("field", form["field"], []string{}, form))
	_, ok = form["field"].(map[string]interface{})
	assert.True(t, ok)
}

func TestValidateJSONConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"json": "{\"str\":\"object\"}",
		},
	}

	set := RuleSet{
		"object":      {"required", "object"},
		"object.json": {"required", "json"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["json"].(map[string]interface{})
	assert.True(t, ok)
}

func TestValidateURL(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateURL("field", "http://www.google.com", []string{}, data))
	assert.True(t, validateURL("field", "https://www.google.com", []string{}, data))
	assert.True(t, validateURL("field", "https://www.google.com?q=a%20surprise%20to%20be%20sure", []string{}, data))
	assert.True(t, validateURL("field", "https://www.google.com/#anchor", []string{}, data))
	assert.True(t, validateURL("field", "https://www.google.com?q=hmm#anchor", []string{}, data))

	assert.False(t, validateURL("field", "https://www.google.com#anchor", []string{}, data))
	assert.False(t, validateURL("field", "www.google.com", []string{}, data))
	assert.False(t, validateURL("field", "w-w.google.com", []string{}, data))
	assert.False(t, validateURL("field", 1, []string{}, data))
	assert.False(t, validateURL("field", 1.2, []string{}, data))
	assert.False(t, validateURL("field", []string{}, []string{}, data))
}

func TestValidateURLConvert(t *testing.T) {
	form := map[string]interface{}{"field": "http://www.google.com"}
	assert.True(t, validateURL("field", form["field"], []string{}, form))
	_, ok := form["field"].(*url.URL)
	assert.True(t, ok)
}

func TestValidateURLConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"url": "http://www.google.com",
		},
	}

	set := RuleSet{
		"object":     {"required", "object"},
		"object.url": {"required", "url"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["url"].(*url.URL)
	assert.True(t, ok)
}

func TestValidateUUID(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateUUID("field", "123e4567-e89b-12d3-a456-426655440000", []string{}, data)) // V1
	assert.True(t, validateUUID("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{}, data)) // V3
	assert.True(t, validateUUID("field", "9125a8dc52ee365ba5aa81b0b3681cf6", []string{}, data))     // V3 no hyphen
	assert.True(t, validateUUID("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{}, data)) // V4
	assert.True(t, validateUUID("field", "11bf5b37e0b842e08dcfdc8c4aefc000", []string{}, data))     // V4 no hyphen
	assert.True(t, validateUUID("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{}, data)) // V5
	assert.True(t, validateUUID("field", "3bbcee75-cecc-5b56-8031-b6641c1ed1f1", []string{}, data)) // V5
	assert.True(t, validateUUID("field", "3bbcee75cecc5b568031b6641c1ed1f1", []string{}, data))     // V5 no hypen

	assert.False(t, validateUUID("field", "hello", []string{}, data))
	assert.False(t, validateUUID("field", 1, []string{}, data))
	assert.False(t, validateUUID("field", 1.2, []string{}, data))
	assert.False(t, validateUUID("field", true, []string{}, data))
	assert.False(t, validateUUID("field", []byte{}, []string{}, data))
}

func TestValidateUUIDConvert(t *testing.T) {
	form := map[string]interface{}{"field": "123e4567-e89b-12d3-a456-426655440000"}
	assert.True(t, validateUUID("field", form["field"], []string{}, form))
	_, ok := form["field"].(uuid.UUID)
	assert.True(t, ok)
}

func TestValidateUUIDConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"uuid": "123e4567-e89b-12d3-a456-426655440000",
		},
	}

	set := RuleSet{
		"object":      {"required", "object"},
		"object.uuid": {"required", "uuid"},
	}

	errors := Validate(data, set, true, "en-US")
	assert.Empty(t, errors)
	_, ok := data["object"].(map[string]interface{})["uuid"].(uuid.UUID)
	assert.True(t, ok)
}

func TestValidateUUIDv3(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateUUID("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"3"}, data))  // V3
	assert.False(t, validateUUID("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"3"}, data)) // V4
	assert.False(t, validateUUID("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"3"}, data)) // V5
}

func TestValidateUUIDv4(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.False(t, validateUUID("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"4"}, data)) // V3
	assert.True(t, validateUUID("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"4"}, data))  // V4
	assert.False(t, validateUUID("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"4"}, data)) // V5
}

func TestValidateUUIDv5(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.False(t, validateUUID("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"5"}, data)) // V3
	assert.False(t, validateUUID("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"5"}, data)) // V4
	assert.True(t, validateUUID("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"5"}, data))  // V5
}
