package config

import (
	"reflect"
	"runtime"
	"testing"
)

func setupConfigBench(b *testing.B) {
	Clear()
	if err := LoadFrom("config.test.json"); err != nil {
		panic(err)
	}
	runtime.GC()
	b.ReportAllocs()
	b.ResetTimer()
}

func BenchmarkValidateInt(b *testing.B) {
	entry := &Entry{1.0, []interface{}{}, reflect.Int, false}
	config := object{"number": entry}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		config.validate("")
		entry.Value = 1.0
	}
}

func BenchmarkSetString(b *testing.B) {
	setupConfigBench(b)
	for n := 0; n < b.N; n++ {
		Set("app.name", "my awesome app")
	}
}

func BenchmarkGet(b *testing.B) {
	setupConfigBench(b)
	for n := 0; n < b.N; n++ {
		Get("app.name")
	}
}

func BenchmarkSetInt(b *testing.B) {
	setupConfigBench(b)
	for n := 0; n < b.N; n++ {
		Set("server.port", 8080)
	}
}
