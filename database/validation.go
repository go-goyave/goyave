package database

import (
	"github.com/System-Glitch/goyave/v3/validation"
)

// This file contains the database-related validation rules

func init() {
	validation.AddRule("unique", &validation.RuleDefinition{
		Function:           validateUnique,
		RequiredParameters: 1,
	})
}

func validateUnique(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	column := field
	if len(parameters) >= 2 {
		column = parameters[1]
	}

	count := int64(0)
	if err := Conn().Table(parameters[0]).Where(column+"= ?", value).Count(&count).Error; err != nil {
		panic(err)
	}
	return count == 0
}
