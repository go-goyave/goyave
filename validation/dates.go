package validation

import (
	"fmt"
	"time"
)

func parseDate(date interface{}, format string) (time.Time, error) {
	str, ok := date.(string)
	if ok {
		t, err := time.Parse(format, str)
		if err == nil {
			return t, err
		}
		return t, err
	}
	return time.Time{}, fmt.Errorf("Date is not a string so cannot be parsed")
}

func getDates(value interface{}, parameters []string, form map[string]interface{}) ([]time.Time, error) {
	dates := []time.Time{}
	date, ok := value.(time.Time)
	if ok {
		dates = append(dates, date)
		for _, param := range parameters {
			_, other, _, exists := GetFieldFromName(param, form)
			if exists {
				otherDate, ok := other.(time.Time)
				if !ok {
					t, err := parseDate(other, "2006-01-02")
					if err != nil {
						return dates, fmt.Errorf("Cannot parse date in other field")
					}
					otherDate = t
				}
				dates = append(dates, otherDate)
				continue
			}

			// TODO v4: avoid reparsing the date every single time
			t, err := parseDate(param, "2006-01-02T15:04:05")
			if err != nil {
				panic(err)
			}
			dates = append(dates, t)
		}

		return dates, nil
	}
	return dates, fmt.Errorf("Value is not a date")
}

func validateDate(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	if len(parameters) == 0 {
		parameters = append(parameters, "2006-01-02")
	}

	t, err := parseDate(value, parameters[0])
	if err == nil {
		// FIXME v4: not ideal because this is done twice. Set parent object in validation context.
		// See: https://github.com/go-goyave/goyave/issues/136
		fieldName, _, parent, _ := GetFieldFromName(field, form)
		parent[fieldName] = t
		return true
	}
	return false
}

func validateBefore(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && dates[0].Before(dates[1])
}

func validateBeforeEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && (dates[0].Before(dates[1]) || dates[0].Equal(dates[1]))
}

func validateAfter(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && dates[0].After(dates[1])
}

func validateAfterEqual(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && (dates[0].After(dates[1]) || dates[0].Equal(dates[1]))
}

func validateDateEquals(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && dates[0].Equal(dates[1])
}

func validateDateBetween(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
	dates, err := getDates(value, parameters, form)
	return err == nil && (dates[0].After(dates[1]) || dates[0].Equal(dates[1])) && (dates[0].Before(dates[2]) || dates[0].Equal(dates[2]))
}
