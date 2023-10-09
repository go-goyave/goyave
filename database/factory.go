package database

import (
	"gorm.io/gorm"
	"goyave.dev/copier"
	"goyave.dev/goyave/v5/util/errors"
)

// Factory an object used to generate records or seed the database.
type Factory[T any] struct {
	generator func() *T
	override  *T
	BatchSize int
}

// NewFactory create a new Factory.
// The given generator function will be used to generate records.
func NewFactory[T any](generator func() *T) *Factory[T] {
	return &Factory[T]{
		generator: generator,
		override:  nil,
		BatchSize: 100,
	}
}

// Override set an override model for generated records.
// Values present in the override model will replace the ones
// in the generated records.
// This function expects a struct pointer as parameter.
// Returns the same instance of `Factory` so this method can be chained.
func (f *Factory[T]) Override(override *T) *Factory[T] {
	f.override = override
	return f
}

// Generate a number of records using the given factory.
func (f *Factory[T]) Generate(count int) []*T {
	if count <= 0 {
		return []*T{}
	}

	slice := make([]*T, 0, count)

	for i := 0; i < count; i++ {
		record := f.generator()
		if f.override != nil {
			if err := copier.CopyWithOption(record, f.override, copier.Option{IgnoreEmpty: true, DeepCopy: true, CaseSensitive: true}); err != nil {
				panic(errors.NewSkip(err, 3))
			}
		}
		slice = append(slice, record)
	}
	return slice
}

// Save generate a number of records using the given factory,
// insert them in the database and return the inserted records.
func (f *Factory[T]) Save(db *gorm.DB, count int) []*T {
	records := f.Generate(count)

	if err := db.CreateInBatches(records, f.BatchSize).Error; err != nil {
		panic(errors.New(err))
	}
	return records
}
