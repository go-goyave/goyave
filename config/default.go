package config

import "reflect"

var configDefaults = object{
	"app": object{
		"name":            &Entry{"goyave", []any{}, reflect.String, false},
		"environment":     &Entry{"localhost", []any{}, reflect.String, false},
		"debug":           &Entry{true, []any{}, reflect.Bool, false},
		"defaultLanguage": &Entry{"en-US", []any{}, reflect.String, false},
	},
	"server": object{
		"host":                  &Entry{"127.0.0.1", []any{}, reflect.String, false},
		"domain":                &Entry{"", []any{}, reflect.String, false},
		"port":                  &Entry{8080, []any{}, reflect.Int, false},
		"writeTimeout":          &Entry{10, []any{}, reflect.Int, false},
		"readTimeout":           &Entry{10, []any{}, reflect.Int, false},
		"idleTimeout":           &Entry{20, []any{}, reflect.Int, false},
		"websocketCloseTimeout": &Entry{10, []any{}, reflect.Int, false},
		"maxUploadSize":         &Entry{10.0, []any{}, reflect.Float64, false},
		"proxy": object{
			"protocol": &Entry{"http", []any{"http", "https"}, reflect.String, false},
			"host":     &Entry{nil, []any{}, reflect.String, false},
			"port":     &Entry{80, []any{}, reflect.Int, false},
			"base":     &Entry{"", []any{}, reflect.String, false},
		},
	},
	"database": object{
		"connection":               &Entry{"none", []any{}, reflect.String, false},
		"host":                     &Entry{"127.0.0.1", []any{}, reflect.String, false},
		"port":                     &Entry{3306, []any{}, reflect.Int, false},
		"name":                     &Entry{"goyave", []any{}, reflect.String, false},
		"username":                 &Entry{"root", []any{}, reflect.String, false},
		"password":                 &Entry{"root", []any{}, reflect.String, false},
		"options":                  &Entry{"charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local", []any{}, reflect.String, false},
		"maxOpenConnections":       &Entry{20, []any{}, reflect.Int, false},
		"maxIdleConnections":       &Entry{20, []any{}, reflect.Int, false},
		"maxLifetime":              &Entry{300, []any{}, reflect.Int, false},
		"defaultReadQueryTimeout":  &Entry{20000, []any{}, reflect.Int, false}, // TODO document this (the timeout set by database.TimeoutPlugin), in ms
		"defaultWriteQueryTimeout": &Entry{40000, []any{}, reflect.Int, false},
		"config": object{
			"skipDefaultTransaction":                   &Entry{false, []any{}, reflect.Bool, false},
			"dryRun":                                   &Entry{false, []any{}, reflect.Bool, false},
			"prepareStmt":                              &Entry{true, []any{}, reflect.Bool, false},
			"disableNestedTransaction":                 &Entry{false, []any{}, reflect.Bool, false},
			"allowGlobalUpdate":                        &Entry{false, []any{}, reflect.Bool, false},
			"disableAutomaticPing":                     &Entry{false, []any{}, reflect.Bool, false},
			"disableForeignKeyConstraintWhenMigrating": &Entry{false, []any{}, reflect.Bool, false},
		},
	},
}

func loadDefaults(src object, dst object) {
	for k, v := range src {
		if obj, ok := v.(object); ok {
			sub := make(object, len(obj))
			loadDefaults(obj, sub)
			dst[k] = sub
		} else {
			entry := v.(*Entry)
			value := entry.Value
			t := reflect.TypeOf(value)
			if t != nil && t.Kind() == reflect.Slice {
				list := reflect.ValueOf(value)
				length := list.Len()
				slice := reflect.MakeSlice(reflect.SliceOf(t.Elem()), 0, length)
				for i := 0; i < length; i++ {
					slice = reflect.Append(slice, list.Index(i))
				}
				value = slice.Interface()
			}
			dst[k] = &Entry{value, entry.AuthorizedValues, entry.Type, entry.IsSlice}
		}
	}
}
