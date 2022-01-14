package httputil

import (
	"regexp"
	"testing"
)

func BenchmarkParseMultiValuesHeader(b *testing.B) {
	multiValuesHeaderRegex = regexp.MustCompile(`^q=([01]\.[0-9]{1,3})$`)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	}
}
