package database

import (
	"reflect"

	"github.com/imdario/mergo"
)

// Generator a generator function generates a single record.
type Generator func() interface{}

// Factory an object used to generate records or seed the database.
type Factory struct {
	generator Generator
	override  interface{}
}

// NewFactory create a new Factory.
// The given generator function will be used to generate records.
func NewFactory(generator Generator) *Factory {
	return &Factory{
		generator: generator,
		override:  nil,
	}
}

// Override set an override model for generated records.
// Values present in the override model will replace the ones
// in the generated records.
// This function expects a struct pointer as parameter.
// Returns the same instance of `Factory` so this method can be chained.
func (f *Factory) Override(override interface{}) *Factory {
	f.override = override
	return f
}

// Generate a number of records using the given factory.
// Returns a slice of the actual type of the generated records,
// meaning you can type-assert safely.
//
//	factory.Generate(5).([]*User)
func (f *Factory) Generate(count int) interface{} {
	if count <= 0 {
		return []interface{}{}
	}
	var t reflect.Type
	var slice reflect.Value
	for i := 0; i < count; i++ {
		record := f.generator()
		if t == nil {
			t = reflect.TypeOf(record)
			slice = reflect.MakeSlice(reflect.SliceOf(t), 0, count)
		}
		if f.override != nil {
			if err := mergo.Merge(record, f.override, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
		slice = reflect.Append(slice, reflect.ValueOf(record))
	}
	return slice.Interface()
}

// Save generate a number of records using the given factory,
// insert them in the database and return the inserted records.
// The returned slice is a slice of the actual type of the generated records,
// meaning you can type-assert safely.
//
//	factory.Save(5).([]*User)
func (f *Factory) Save(count int) interface{} {
	db := GetConnection()
	records := f.Generate(count)

	if err := db.Create(records).Error; err != nil {
		panic(err)
	}
	return records
}
