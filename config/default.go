package config

import "reflect"

var configDefaults = object{
	"app": object{
		"name":            &Entry{"goyave", []any{}, reflect.String, false, true},
		"environment":     &Entry{"localhost", []any{}, reflect.String, false, true},
		"debug":           &Entry{true, []any{}, reflect.Bool, false, true},
		"defaultLanguage": &Entry{"en-US", []any{}, reflect.String, false, true},
	},
	"server": object{
		"host":                  &Entry{"127.0.0.1", []any{}, reflect.String, false, true},
		"domain":                &Entry{"", []any{}, reflect.String, false, true},
		"port":                  &Entry{8080, []any{}, reflect.Int, false, true},
		"writeTimeout":          &Entry{10, []any{}, reflect.Int, false, true},
		"readTimeout":           &Entry{10, []any{}, reflect.Int, false, true},
		"readHeaderTimeout":     &Entry{10, []any{}, reflect.Int, false, true},
		"idleTimeout":           &Entry{20, []any{}, reflect.Int, false, true},
		"websocketCloseTimeout": &Entry{10, []any{}, reflect.Int, false, true},
		"maxUploadSize":         &Entry{10.0, []any{}, reflect.Float64, false, true},
		"proxy": object{
			"protocol": &Entry{"http", []any{"http", "https"}, reflect.String, false, true},
			"host":     &Entry{nil, []any{}, reflect.String, false, false},
			"port":     &Entry{80, []any{}, reflect.Int, false, true},
			"base":     &Entry{"", []any{}, reflect.String, false, true},
		},
	},
	"database": object{
		"connection":               &Entry{"none", []any{}, reflect.String, false, true},
		"host":                     &Entry{"127.0.0.1", []any{}, reflect.String, false, true},
		"port":                     &Entry{0, []any{}, reflect.Int, false, true},
		"name":                     &Entry{"", []any{}, reflect.String, false, true},
		"username":                 &Entry{"", []any{}, reflect.String, false, true},
		"password":                 &Entry{"", []any{}, reflect.String, false, true},
		"options":                  &Entry{"", []any{}, reflect.String, false, true},
		"maxOpenConnections":       &Entry{20, []any{}, reflect.Int, false, true},
		"maxIdleConnections":       &Entry{20, []any{}, reflect.Int, false, true},
		"maxLifetime":              &Entry{300, []any{}, reflect.Int, false, true},
		"defaultReadQueryTimeout":  &Entry{20000, []any{}, reflect.Int, false, true},
		"defaultWriteQueryTimeout": &Entry{40000, []any{}, reflect.Int, false, true},
		"config": object{
			"skipDefaultTransaction":                   &Entry{false, []any{}, reflect.Bool, false, true},
			"dryRun":                                   &Entry{false, []any{}, reflect.Bool, false, true},
			"prepareStmt":                              &Entry{true, []any{}, reflect.Bool, false, true},
			"disableNestedTransaction":                 &Entry{false, []any{}, reflect.Bool, false, true},
			"allowGlobalUpdate":                        &Entry{false, []any{}, reflect.Bool, false, true},
			"disableAutomaticPing":                     &Entry{false, []any{}, reflect.Bool, false, true},
			"disableForeignKeyConstraintWhenMigrating": &Entry{false, []any{}, reflect.Bool, false, true},
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
			dst[k] = &Entry{value, entry.AuthorizedValues, entry.Type, entry.IsSlice, entry.Required}
		}
	}
}
