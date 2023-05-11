package database

import (
	"goyave.dev/goyave/v4/validation"
)

// This file contains the database-related validation rules

func init() {
	validation.AddRule("unique", &validation.RuleDefinition{
		Function:           validateUnique,
		RequiredParameters: 1,
	})
	validation.AddRule("exists", &validation.RuleDefinition{
		Function:           validateExists,
		RequiredParameters: 1,
	})
}

func validateUnique(ctx *validation.Context) bool {
	if !ctx.Valid() {
		return true
	}
	column := ctx.Name
	if len(ctx.Rule.Params) >= 2 {
		column = ctx.Rule.Params[1]
	}

	count := int64(0)
	if err := Conn().Table(ctx.Rule.Params[0]).Where(column+"= ?", ctx.Value).Count(&count).Error; err != nil {
		panic(err)
	}
	return count == 0
}

func validateExists(ctx *validation.Context) bool {
	return !validateUnique(ctx)
}
