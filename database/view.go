package database

// IView models implementing this interface are identified as SQL views
// if IsView() returns true.
// Because records cannot be deleted from views, this is useful in test suites
// so `ClearDatabase()` doesn't try (and fails) to delete the records.
type IView interface {
	IsView() bool
}

// View helper implementing IView. Useful for composition inside models rather
// than implementing IView manually.
type View struct{}

// IsView always returns true.
func (v View) IsView() bool {
	return true
}
