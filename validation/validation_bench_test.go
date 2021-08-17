package validation

import (
	"runtime"
	"testing"
)

func setupValidationBench(b *testing.B) {
	b.ReportAllocs()
	runtime.GC()
	b.ResetTimer()
}

func BenchmarkValidateWithParsing(b *testing.B) {
	set := RuleSet{
		"email":    {"required", "string", "between:3,125", "email"},
		"password": {"required", "string", "between:6,64", "confirmed"},
		"info":     {"nullable", "array:string", ">min:2"},
	}
	data := map[string]interface{}{
		"email":                 "pedro@example.org",
		"password":              "this is a strong password",
		"password_confirmation": "this is a strong password",
		"info":                  []string{"smart", "reliable"},
	}
	setupValidationBench(b)
	for n := 0; n < b.N; n++ {
		Validate(data, set, true, "en-US")
	}
}

func BenchmarkValidatePreParsed(b *testing.B) {
	rules := &Rules{
		Fields: FieldMap{
			"email": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
					{Name: "between", Params: []string{"3", "125"}},
					{Name: "email"},
				},
			},
			"password": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
					{Name: "between", Params: []string{"6", "64"}},
					{Name: "confirmed"},
				},
			},
			"info": {
				Rules: []*Rule{
					{Name: "nullable"},
					{Name: "array", Params: []string{"string"}},
					{Name: "min", Params: []string{"2"}, ArrayDimension: 1},
				},
			},
		},
	}
	rules.Check()
	data := map[string]interface{}{
		"email":                 "pedro@example.org",
		"password":              "this is a strong password",
		"password_confirmation": "this is a strong password",
		"info":                  []string{"smart", "reliable"},
	}

	setupValidationBench(b)
	for n := 0; n < b.N; n++ {
		Validate(data, rules, true, "en-US")
	}
}

func BenchmarkParseAndCheck(b *testing.B) {
	set := RuleSet{
		"email":    {"required", "string", "between:3,125", "email"},
		"password": {"required", "string", "between:6,64", "confirmed"},
		"info":     {"nullable", "array:string", ">min:2"},
	}
	setupValidationBench(b)
	for n := 0; n < b.N; n++ {
		set.parse()
	}
}

func BenchmarkCheck(b *testing.B) {
	rules := &Rules{
		Fields: FieldMap{
			"email": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
					{Name: "between", Params: []string{"3", "125"}},
					{Name: "email"},
				},
			},
			"password": {
				Rules: []*Rule{
					{Name: "required"},
					{Name: "string"},
					{Name: "between", Params: []string{"6", "64"}},
					{Name: "confirmed"},
				},
			},
			"info": {
				Rules: []*Rule{
					{Name: "nullable"},
					{Name: "array", Params: []string{"string"}},
					{Name: "min", Params: []string{"2"}, ArrayDimension: 1},
				},
			},
		},
	}
	setupValidationBench(b)
	for n := 0; n < b.N; n++ {
		rules.Check()
		rules.checked = false
	}
}
