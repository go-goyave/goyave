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
	assert.True(t, validateString(newTestContext("field", "string", []string{}, map[string]interface{}{})))
	assert.False(t, validateString(newTestContext("field", 2, []string{}, map[string]interface{}{})))
	assert.False(t, validateString(newTestContext("field", 2.5, []string{}, map[string]interface{}{})))
	assert.False(t, validateString(newTestContext("field", []byte{}, []string{}, map[string]interface{}{})))
	assert.False(t, validateString(newTestContext("field", []string{}, []string{}, map[string]interface{}{})))
}

func TestValidateDigits(t *testing.T) {
	assert.True(t, validateDigits(newTestContext("field", "123", []string{}, map[string]interface{}{})))
	assert.True(t, validateDigits(newTestContext("field", "0123456789", []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", "2.3", []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", "-123", []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", "abcd", []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", "/*-&é\"'(-è_ç", []string{}, map[string]interface{}{})))

	// Not string
	assert.False(t, validateDigits(newTestContext("field", 1, []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", 1.2, []string{}, map[string]interface{}{})))
	assert.False(t, validateDigits(newTestContext("field", true, []string{}, map[string]interface{}{})))
}

func TestValidateRegex(t *testing.T) {
	assert.True(t, validateRegex(newTestContext("field", "sghtyhg", []string{"t"}, map[string]interface{}{})))
	assert.True(t, validateRegex(newTestContext("field", "sghtyhg", []string{"[^\\s]"}, map[string]interface{}{})))
	assert.False(t, validateRegex(newTestContext("field", "sgh tyhg", []string{"^[^\\s]+$"}, map[string]interface{}{})))
	assert.False(t, validateRegex(newTestContext("field", "48s9", []string{"^[^0-9]+$"}, map[string]interface{}{})))
	assert.True(t, validateRegex(newTestContext("field", "489", []string{"^[0-9]+$"}, map[string]interface{}{})))
	assert.False(t, validateRegex(newTestContext("field", 489, []string{"^[^0-9]+$"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		validateRegex(newTestContext("field", "", []string{"doesn't compile \\"}, map[string]interface{}{}))
	})

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "regex"},
			},
		}
		field.Check()
	})
}

func TestValidateEmail(t *testing.T) {
	assert.True(t, validateEmail(newTestContext("field", "simple@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "very.common@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "disposable.style.email.with+symbol@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "other.email-with-hyphen@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "fully-qualified-domain@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "user.name+tag+sorting@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "x@example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "example-indeed@strange-example.com", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "admin@mailserver1", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "example@s.example", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "\" \"@example.org", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "\"john..doe\"@example.org", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "mailhost!username@example.org", []string{}, map[string]interface{}{})))
	assert.True(t, validateEmail(newTestContext("field", "user%example.com@example.org", []string{}, map[string]interface{}{})))
	assert.False(t, validateEmail(newTestContext("field", "Abc.example.com", []string{}, map[string]interface{}{})))
	assert.False(t, validateEmail(newTestContext("field", "1234567890123456789012345678901234567890123456789012345678901234+x@example.com", []string{}, map[string]interface{}{})))
}

func TestValidateAlpha(t *testing.T) {
	assert.True(t, validateAlpha(newTestContext("field", "helloworld", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlpha(newTestContext("field", "éèçàû", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlpha(newTestContext("field", "hello world", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlpha(newTestContext("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlpha(newTestContext("field", "helloworld2", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlpha(newTestContext("field", 2, []string{}, map[string]interface{}{})))
}

func TestValidateAlphaDash(t *testing.T) {
	assert.True(t, validateAlphaDash(newTestContext("field", "helloworld", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlphaDash(newTestContext("field", "éèçàû_-", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlphaDash(newTestContext("field", "hello-world", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlphaDash(newTestContext("field", "hello-world_2", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaDash(newTestContext("field", "hello world", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaDash(newTestContext("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaDash(newTestContext("field", 2, []string{}, map[string]interface{}{})))
}

func TestValidateAlphaNumeric(t *testing.T) {
	assert.True(t, validateAlphaNumeric(newTestContext("field", "helloworld2", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlphaNumeric(newTestContext("field", "éèçàû2", []string{}, map[string]interface{}{})))
	assert.True(t, validateAlphaNumeric(newTestContext("field", "helloworld2", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaNumeric(newTestContext("field", "hello world", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaNumeric(newTestContext("field", "/+*(@)={}\"'", []string{}, map[string]interface{}{})))
	assert.False(t, validateAlphaNumeric(newTestContext("field", 2, []string{}, map[string]interface{}{})))
}

func TestValidateStartsWith(t *testing.T) {
	assert.True(t, validateStartsWith(newTestContext("field", "hello world", []string{"hello"}, map[string]interface{}{})))
	assert.True(t, validateStartsWith(newTestContext("field", "hi", []string{"hello", "hi", "hey"}, map[string]interface{}{})))
	assert.False(t, validateStartsWith(newTestContext("field", "sup'!", []string{"hello", "hi", "hey"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "starts_with"},
			},
		}
		field.Check()
	})
}

func TestValidateEndsWith(t *testing.T) {
	assert.True(t, validateEndsWith(newTestContext("field", "hello world", []string{"world"}, map[string]interface{}{})))
	assert.True(t, validateEndsWith(newTestContext("field", "oh hi mark", []string{"ross", "mark", "bruce"}, map[string]interface{}{})))
	assert.False(t, validateEndsWith(newTestContext("field", "sup' bro!", []string{"ross", "mark", "bruce"}, map[string]interface{}{})))

	assert.Panics(t, func() {
		field := &Field{
			Rules: []*Rule{
				{Name: "ends_with"},
			},
		}
		field.Check()
	})
}

func TestValidateTimezone(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.True(t, validateTimezone(newTestContext("field", "UTC", []string{}, data)))
	assert.True(t, validateTimezone(newTestContext("field", "Europe/Paris", []string{}, data)))
	assert.True(t, validateTimezone(newTestContext("field", "America/St_Thomas", []string{}, data)))
	assert.True(t, validateTimezone(newTestContext("field", "GMT", []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", "GMT+2", []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", "UTC+2", []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", "here", []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", 1.5, []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", true, []string{}, data)))
	assert.False(t, validateTimezone(newTestContext("field", []string{"UTC"}, []string{}, data)))
}

func TestValidateTimezoneConvert(t *testing.T) {
	form := map[string]interface{}{"field": "UTC"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateTimezone(ctx))

	_, ok := ctx.Value.(*time.Location)
	assert.True(t, ok)
}

func TestValidateTimezoneConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"timezone": "UTC",
		},
	}

	set := RuleSet{
		"object":          List{"required", "object"},
		"object.timezone": List{"required", "timezone"},
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
	assert.True(t, validateIP(newTestContext("field", "127.0.0.1", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "192.168.0.1", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "88.88.88.88", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data)))
	assert.True(t, validateIP(newTestContext("field", "::1", []string{}, data)))

	assert.False(t, validateIP(newTestContext("field", "1", []string{}, data)))
	assert.False(t, validateIP(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateIP(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateIP(newTestContext("field", true, []string{}, data)))
	assert.False(t, validateIP(newTestContext("field", []byte{}, []string{}, data)))
}

func TestValidateIPConvert(t *testing.T) {
	form := map[string]interface{}{"field": "127.0.0.1"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateIP(ctx))

	_, ok := ctx.Value.(net.IP)
	assert.True(t, ok)
}

func TestValidateIPConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "127.0.0.1",
		},
	}

	set := RuleSet{
		"object":    List{"required", "object"},
		"object.ip": List{"required", "ip"},
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
	assert.True(t, validateIPv4(newTestContext("field", "127.0.0.1", []string{}, data)))
	assert.True(t, validateIPv4(newTestContext("field", "192.168.0.1", []string{}, data)))
	assert.True(t, validateIPv4(newTestContext("field", "88.88.88.88", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", "::1", []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", true, []string{}, data)))
	assert.False(t, validateIPv4(newTestContext("field", []byte{}, []string{}, data)))
}

func TestValidateIPv4ConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "127.0.0.1",
		},
	}

	set := RuleSet{
		"object":    List{"required", "object"},
		"object.ip": List{"required", "ipv4"},
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
	assert.False(t, validateIPv6(newTestContext("field", "127.0.0.1", []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", "192.168.0.1", []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", "88.88.88.88", []string{}, data)))
	assert.True(t, validateIPv6(newTestContext("field", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", []string{}, data)))
	assert.True(t, validateIPv6(newTestContext("field", "2001:db8:85a3::8a2e:370:7334", []string{}, data)))
	assert.True(t, validateIPv6(newTestContext("field", "2001:db8:85a3:0:0:8a2e:370:7334", []string{}, data)))
	assert.True(t, validateIPv6(newTestContext("field", "2001:db8:85a3:8d3:1319:8a2e:370:7348", []string{}, data)))
	assert.True(t, validateIPv6(newTestContext("field", "::1", []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", true, []string{}, data)))
	assert.False(t, validateIPv6(newTestContext("field", []byte{}, []string{}, data)))
}

func TestValidateIPv6ConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"ip": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		},
	}

	set := RuleSet{
		"object":    List{"required", "object"},
		"object.ip": List{"required", "ipv6"},
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
	assert.True(t, validateJSON(newTestContext("field", "2", []string{}, data)))
	assert.True(t, validateJSON(newTestContext("field", "2.5", []string{}, data)))
	assert.True(t, validateJSON(newTestContext("field", "\"str\"", []string{}, data)))
	assert.True(t, validateJSON(newTestContext("field", "[\"str\",\"array\"]", []string{}, data)))
	assert.True(t, validateJSON(newTestContext("field", "{\"str\":\"object\"}", []string{}, data)))
	assert.True(t, validateJSON(newTestContext("field", "{\"str\":[\"object\",\"array\"]}", []string{}, data)))

	assert.False(t, validateJSON(newTestContext("field", "{str:[\"object\",\"array\"]}", []string{}, data)))
	assert.False(t, validateJSON(newTestContext("field", "", []string{}, data)))
	assert.False(t, validateJSON(newTestContext("field", "\"d", []string{}, data)))
	assert.False(t, validateJSON(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateJSON(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateJSON(newTestContext("field", map[string]string{}, []string{}, data)))
}

func TestValidateJSONConvert(t *testing.T) {
	form := map[string]interface{}{"field": "2"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateJSON(ctx))
	_, ok := ctx.Value.(float64)
	assert.True(t, ok)

	form = map[string]interface{}{"field": "\"str\""}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateJSON(ctx))
	_, ok = ctx.Value.(string)
	assert.True(t, ok)

	form = map[string]interface{}{"field": "[\"str\",\"array\"]"}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateJSON(ctx))
	_, ok = ctx.Value.([]interface{})
	assert.True(t, ok)

	form = map[string]interface{}{"field": "{\"str\":\"object\"}"}
	ctx = newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateJSON(ctx))
	_, ok = ctx.Value.(map[string]interface{})
	assert.True(t, ok)
}

func TestValidateJSONConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"json": "{\"str\":\"object\"}",
		},
	}

	set := RuleSet{
		"object":      List{"required", "object"},
		"object.json": List{"required", "json"},
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
	assert.True(t, validateURL(newTestContext("field", "http://www.google.com", []string{}, data)))
	assert.True(t, validateURL(newTestContext("field", "https://www.google.com", []string{}, data)))
	assert.True(t, validateURL(newTestContext("field", "https://www.google.com?q=a%20surprise%20to%20be%20sure", []string{}, data)))
	assert.True(t, validateURL(newTestContext("field", "https://www.google.com/#anchor", []string{}, data)))
	assert.True(t, validateURL(newTestContext("field", "https://www.google.com?q=hmm#anchor", []string{}, data)))

	assert.False(t, validateURL(newTestContext("field", "https://www.google.com#anchor", []string{}, data)))
	assert.False(t, validateURL(newTestContext("field", "www.google.com", []string{}, data)))
	assert.False(t, validateURL(newTestContext("field", "w-w.google.com", []string{}, data)))
	assert.False(t, validateURL(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateURL(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateURL(newTestContext("field", []string{}, []string{}, data)))
}

func TestValidateURLConvert(t *testing.T) {
	form := map[string]interface{}{"field": "http://www.google.com"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateURL(ctx))
	_, ok := ctx.Value.(*url.URL)
	assert.True(t, ok)
}

func TestValidateURLConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"url": "http://www.google.com",
		},
	}

	set := RuleSet{
		"object":     List{"required", "object"},
		"object.url": List{"required", "url"},
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
	assert.True(t, validateUUID(newTestContext("field", "123e4567-e89b-12d3-a456-426655440000", []string{}, data))) // V1
	assert.True(t, validateUUID(newTestContext("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{}, data))) // V3
	assert.True(t, validateUUID(newTestContext("field", "9125a8dc52ee365ba5aa81b0b3681cf6", []string{}, data)))     // V3 no hyphen
	assert.True(t, validateUUID(newTestContext("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{}, data))) // V4
	assert.True(t, validateUUID(newTestContext("field", "11bf5b37e0b842e08dcfdc8c4aefc000", []string{}, data)))     // V4 no hyphen
	assert.True(t, validateUUID(newTestContext("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{}, data))) // V5
	assert.True(t, validateUUID(newTestContext("field", "3bbcee75-cecc-5b56-8031-b6641c1ed1f1", []string{}, data))) // V5
	assert.True(t, validateUUID(newTestContext("field", "3bbcee75cecc5b568031b6641c1ed1f1", []string{}, data)))     // V5 no hypen

	assert.False(t, validateUUID(newTestContext("field", "hello", []string{}, data)))
	assert.False(t, validateUUID(newTestContext("field", 1, []string{}, data)))
	assert.False(t, validateUUID(newTestContext("field", 1.2, []string{}, data)))
	assert.False(t, validateUUID(newTestContext("field", true, []string{}, data)))
	assert.False(t, validateUUID(newTestContext("field", []byte{}, []string{}, data)))
}

func TestValidateUUIDConvert(t *testing.T) {
	form := map[string]interface{}{"field": "123e4567-e89b-12d3-a456-426655440000"}
	ctx := newTestContext("field", form["field"], []string{}, form)
	assert.True(t, validateUUID(ctx))
	_, ok := ctx.Value.(uuid.UUID)
	assert.True(t, ok)
}

func TestValidateUUIDConvertInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"uuid": "123e4567-e89b-12d3-a456-426655440000",
		},
	}

	set := RuleSet{
		"object":      List{"required", "object"},
		"object.uuid": List{"required", "uuid"},
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
	assert.True(t, validateUUID(newTestContext("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"3"}, data)))  // V3
	assert.False(t, validateUUID(newTestContext("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"3"}, data))) // V4
	assert.False(t, validateUUID(newTestContext("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"3"}, data))) // V5
}

func TestValidateUUIDv4(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.False(t, validateUUID(newTestContext("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"4"}, data))) // V3
	assert.True(t, validateUUID(newTestContext("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"4"}, data)))  // V4
	assert.False(t, validateUUID(newTestContext("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"4"}, data))) // V5
}

func TestValidateUUIDv5(t *testing.T) {
	data := map[string]interface{}{
		"field": "",
	}
	assert.False(t, validateUUID(newTestContext("field", "9125a8dc-52ee-365b-a5aa-81b0b3681cf6", []string{"5"}, data))) // V3
	assert.False(t, validateUUID(newTestContext("field", "11bf5b37-e0b8-42e0-8dcf-dc8c4aefc000", []string{"5"}, data))) // V4
	assert.True(t, validateUUID(newTestContext("field", "fdda765f-fc57-5604-a269-52a7df8164ec", []string{"5"}, data)))  // V5
}
