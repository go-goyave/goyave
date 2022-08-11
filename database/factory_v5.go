package database

import (
	"github.com/imdario/mergo"
	"gorm.io/gorm"
)

// Factory an object used to generate records or seed the database.
type FactoryV5[T any] struct {
	db        *gorm.DB
	generator func() *T
	override  *T
}

// NewFactory create a new Factory.
// The given generator function will be used to generate records.
func NewFactoryV5[T any](generator func() *T) *FactoryV5[T] {
	return &FactoryV5[T]{
		generator: generator,
		override:  nil,
	}
}

// Override set an override model for generated records.
// Values present in the override model will replace the ones
// in the generated records.
// This function expects a struct pointer as parameter.
// Returns the same instance of `Factory` so this method can be chained.
func (f *FactoryV5[T]) Override(override *T) *FactoryV5[T] {
	f.override = override
	return f
}

// Generate a number of records using the given factory.
func (f *FactoryV5[T]) Generate(count int) []*T {
	if count <= 0 {
		return []*T{}
	}

	slice := make([]*T, 0, count)

	for i := 0; i < count; i++ {
		record := f.generator()
		if f.override != nil {
			if err := mergo.Merge(record, f.override, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
		slice = append(slice, record)
	}
	return slice
}

// Save generate a number of records using the given factory,
// insert them in the database and return the inserted records.
func (f *FactoryV5[T]) Save(db *gorm.DB, count int) []*T {
	records := f.Generate(count)

	if err := db.Create(records).Error; err != nil {
		panic(err)
	}
	return records
}
