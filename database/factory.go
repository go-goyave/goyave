package database

import "github.com/imdario/mergo"

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
func (f *Factory) Override(override interface{}) *Factory {
	f.override = override
	return f
}

// Generate a number of records using the given factory.
func (f *Factory) Generate(count uint) []interface{} {
	records := make([]interface{}, 0, count)
	for i := uint(0); i < count; i++ {
		record := f.generator()
		if f.override != nil {
			if err := mergo.Merge(record, f.override, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
		records = append(records, record)
	}
	return records
}

// Save generate a number of records using the given factory and return
func (f *Factory) Save(count uint) {
	// TODO bulk insert for better performance

	db := GetConnection()
	for _, record := range f.Generate(count) {
		db.Create(record)
	}
}
