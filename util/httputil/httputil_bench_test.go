package httputil

import (
	"testing"
)

func BenchmarkParseMultiValuesHeader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	}
}
