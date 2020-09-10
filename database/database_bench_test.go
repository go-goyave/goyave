package database

import (
	"runtime"
	"testing"

	"github.com/System-Glitch/goyave/v3/config"
	"gorm.io/driver/mysql"
)

func setupDatabaseBench(b *testing.B) {
	if err := config.LoadFrom("config.test.json"); err != nil {
		panic(err)
	}
	runtime.GC()
	b.ReportAllocs()
	b.ResetTimer()
}

func BenchmarkBuildConnectionOptions(b *testing.B) {
	d := dialect{"{username}:{password}@({host}:{port})/{name}?{options}", mysql.Open}
	setupDatabaseBench(b)
	for n := 0; n < b.N; n++ {
		d.buildDSN()
	}
}
