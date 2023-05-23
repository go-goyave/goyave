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
func (v *UUIDValidator) Validate(ctx *Context) bool {
	if uid, ok := ctx.Value.(uuid.UUID); ok {
		return v.checkVersion(uid)
	}
	val, ok := ctx.Value.(string)
	if !ok {
		return false
	}
	uid, err := uuid.Parse(val)
	if err != nil {
		return false
	}

	ok = v.checkVersion(uid)
	if ok {
		ctx.Value = uid
	}
	return ok
}

func (v *UUIDValidator) checkVersion(uid uuid.UUID) bool {
	return len(v.AcceptedVersions) == 0 || lo.Contains(v.AcceptedVersions, uid.Version())
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
