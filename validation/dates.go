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

func getDates(ctx *Context) ([]time.Time, error) {
	dates := []time.Time{}
	date, ok := ctx.Value.(time.Time)
	if ok {
		dates = append(dates, date)
		for _, param := range ctx.Rule.Params {
			_, other, _, exists := GetFieldFromName(param, ctx.Data)
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

func validateDate(ctx *Context) bool {
	if len(ctx.Rule.Params) == 0 {
		ctx.Rule.Params = append(ctx.Rule.Params, "2006-01-02")
	}

	t, err := parseDate(ctx.Value, ctx.Rule.Params[0])
	if err == nil {
		ctx.Value = t
		return true
	}
	return false
}

func validateBefore(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && dates[0].Before(dates[1])
}

func validateBeforeEqual(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && (dates[0].Before(dates[1]) || dates[0].Equal(dates[1]))
}

func validateAfter(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && dates[0].After(dates[1])
}

func validateAfterEqual(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && (dates[0].After(dates[1]) || dates[0].Equal(dates[1]))
}

func validateDateEquals(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && dates[0].Equal(dates[1])
}

func validateDateBetween(ctx *Context) bool {
	dates, err := getDates(ctx)
	return err == nil && (dates[0].After(dates[1]) || dates[0].Equal(dates[1])) && (dates[0].Before(dates[2]) || dates[0].Equal(dates[2]))
}

func validateDateBeforeNow(ctx *Context) bool {
	if date, ok := ctx.Value.(time.Time); ok {
		return date.Before(ctx.Now)
	}
	return false
}

func validateDateAfterNow(ctx *Context) bool {
	if date, ok := ctx.Value.(time.Time); ok {
		return date.After(ctx.Now)
	}
	return false
}
