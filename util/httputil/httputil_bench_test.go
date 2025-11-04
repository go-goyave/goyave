package httputil

import (
	"testing"
)

func BenchmarkParseMultiValuesHeader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	}
}
