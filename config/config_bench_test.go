package config

import (
	"reflect"
	"runtime"
	"testing"
)

func setupConfigBench(b *testing.B) {
	Clear()
	if err := LoadFrom("config.bench.json"); err != nil {
		panic(err)
	}
	runtime.GC()
	b.ReportAllocs()
	b.ResetTimer()
}

func setupConfigBenchV5(b *testing.B) *Config {
	cfg, err := LoadFromV5("config.bench.json")
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	defer b.ResetTimer()
	return cfg
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

func BenchmarkGetString(b *testing.B) {
	setupConfigBench(b)
	for n := 0; n < b.N; n++ {
		GetString("app.name")
	}
}

func BenchmarkGetStringParallel(b *testing.B) {
	setupConfigBench(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetString("app.name")
		}
	})
}

func BenchmarkSetInt(b *testing.B) {
	setupConfigBench(b)
	for n := 0; n < b.N; n++ {
		Set("server.port", 8080)
	}
}

func BenchmarkGetStringV5(b *testing.B) {
	cfg := setupConfigBenchV5(b)
	for n := 0; n < b.N; n++ {
		cfg.GetString("app.name")
	}
}

func BenchmarkGetStringParallelV5(b *testing.B) { // It yields much better results than when using a RWMutex
	cfg := setupConfigBenchV5(b)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cfg.GetString("app.name")
		}
	})
}
