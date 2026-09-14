package typeutil

import (
	"testing"
)

func BenchmarkConvert(b *testing.B) {
	in := map[string]any{"p": "p", "a": "hello", "b": 0.3, "d": []string{"world"}, "c": 456, "nested": map[string]any{"c": 123}}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		out, _ := Convert[*TestStruct](in)
		_ = out
	}
}
