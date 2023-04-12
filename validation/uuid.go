package validation

import (
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// UUIDValidator the field under validation must be a string representing
// a valid UUID.
// If one or more `accepterVersions` are provided, the parsed UUID must
// be a UUID of one of these versions. If none are given, all versions are
// accepted.
//
// If validation passes, the value is converted to `uuid.UUID`.
type UUIDValidator struct {
	BaseValidator
	AcceptedVersions []uuid.Version
}

// Validate checks the field under validation satisfies this validator's criteria.
func (v *UUIDValidator) Validate(ctx *ContextV5) bool {
	if _, ok := ctx.Value.(uuid.UUID); ok {
		return true
	}
	val, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return false
	}

	ok = len(v.AcceptedVersions) == 0 || lo.Contains(v.AcceptedVersions, id.Version())
	if ok {
		ctx.Value = id
	}
	return ok
}

// Name returns the string name of the validator.
func (v *UUIDValidator) Name() string { return "uuid" }

// IsType returns true.
func (v *UUIDValidator) IsType() bool { return true }

// TODO specify accepted versions in validation message?

// UUID the field under validation must be a string representing
// a valid UUID.
// If one or more `accepterVersions` are provided, the parsed UUID must
// be a UUID of one of these versions. If none are given, all versions are
// accepted.
//
// If validation passes, the value is converted to `uuid.UUID`.
func UUID(acceptedVersions ...uuid.Version) *UUIDValidator {
	return &UUIDValidator{AcceptedVersions: acceptedVersions}
}
