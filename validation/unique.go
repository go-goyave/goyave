package validation

import "gorm.io/gorm"

// UniqueValidator validates the field under validation must have a unique value in database
// according to the provided database scope. Uniqueness is checked using a COUNT query.
type UniqueValidator struct {
	BaseValidator
	Scope func(db *gorm.DB, val any) *gorm.DB
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *UniqueValidator) Validate(ctx *ContextV5) bool {
	if !ctx.Valid() {
		return true
	}
	count := int64(0)

	if err := v.Scope(v.DB(), ctx.Value).Count(&count).Error; err != nil {
		ctx.AddError(err)
		return false
	}
	return count == 0
}

// Name returns the string name of the validator.
func (v *UniqueValidator) Name() string { return "unique" }

// IsTypeDependent returns true.
func (v *UniqueValidator) IsTypeDependent() bool { return true }

// Unique validates the field under validation must have a unique value in database
// according to the provided database scope. Uniqueness is checked using a COUNT query.
//
//	 v.Unique(func(db *gorm.DB, val any) *gorm.DB {
//		return db.Model(&model.User{}).Where(clause.PrimaryKey, val)
//	 })
//
//	 v.Unique(func(db *gorm.DB, val any) *gorm.DB {
//		// Unique email excluding the currently authenticated user
//		return db.Model(&model.User{}).Where("email", val).Where("email != ?", request.User.(*model.User).Email)
//	 })
func Unique(scope func(db *gorm.DB, val any) *gorm.DB) *UniqueValidator {
	return &UniqueValidator{Scope: scope}
}

//------------------------------

// ExistsValidator validates the field under validation must exist in database
// according to the provided database scope. Existence is checked using a COUNT query.
type ExistsValidator struct {
	UniqueValidator
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *ExistsValidator) Validate(ctx *ContextV5) bool {
	return !v.UniqueValidator.Validate(ctx)
}

// Name returns the string name of the validator.
func (v *ExistsValidator) Name() string { return "exists" }

// Exists validates the field under validation must have exist database
// according to the provided database scope. Existence is checked using a COUNT query.
//
//	 v.Exists(func(db *gorm.DB, val any) *gorm.DB {
//		return db.Model(&model.User{}).Where(clause.PrimaryKey, val)
//	 })
func Exists(scope func(db *gorm.DB, val any) *gorm.DB) *ExistsValidator {
	return &ExistsValidator{UniqueValidator: UniqueValidator{Scope: scope}}
}
