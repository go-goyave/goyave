package config

import "reflect"

var configDefaults = object{
	"app": object{
		"name":            &Entry{"goyave", []any{}, reflect.String, false, false},
		"environment":     &Entry{"localhost", []any{}, reflect.String, false, false},
		"debug":           &Entry{true, []any{}, reflect.Bool, false, false},
		"defaultLanguage": &Entry{"en-US", []any{}, reflect.String, false, false},
	},
	"server": object{
		"host":                  &Entry{"127.0.0.1", []any{}, reflect.String, false, false},
		"domain":                &Entry{"", []any{}, reflect.String, false, false},
		"port":                  &Entry{8080, []any{}, reflect.Int, false, false},
		"writeTimeout":          &Entry{10, []any{}, reflect.Int, false, false},
		"readTimeout":           &Entry{10, []any{}, reflect.Int, false, false},
		"readHeaderTimeout":     &Entry{10, []any{}, reflect.Int, false, false},
		"idleTimeout":           &Entry{20, []any{}, reflect.Int, false, false},
		"websocketCloseTimeout": &Entry{10, []any{}, reflect.Int, false, false},
		"maxUploadSize":         &Entry{10.0, []any{}, reflect.Float64, false, false},
		"proxy": object{
			"protocol": &Entry{"http", []any{"http", "https"}, reflect.String, false, false},
			"host":     &Entry{nil, []any{}, reflect.String, false, false},
			"port":     &Entry{80, []any{}, reflect.Int, false, false},
			"base":     &Entry{"", []any{}, reflect.String, false, false},
		},
	},
	"database": object{
		"connection":               &Entry{"none", []any{}, reflect.String, false, false},
		"host":                     &Entry{"127.0.0.1", []any{}, reflect.String, false, false},
		"port":                     &Entry{0, []any{}, reflect.Int, false, false},
		"name":                     &Entry{"", []any{}, reflect.String, false, false},
		"username":                 &Entry{"", []any{}, reflect.String, false, false},
		"password":                 &Entry{"", []any{}, reflect.String, false, false},
		"options":                  &Entry{"", []any{}, reflect.String, false, false},
		"maxOpenConnections":       &Entry{20, []any{}, reflect.Int, false, false},
		"maxIdleConnections":       &Entry{20, []any{}, reflect.Int, false, false},
		"maxLifetime":              &Entry{300, []any{}, reflect.Int, false, false},
		"defaultReadQueryTimeout":  &Entry{20000, []any{}, reflect.Int, false, false},
		"defaultWriteQueryTimeout": &Entry{40000, []any{}, reflect.Int, false, false},
		"config": object{
			"skipDefaultTransaction":                   &Entry{false, []any{}, reflect.Bool, false, false},
			"dryRun":                                   &Entry{false, []any{}, reflect.Bool, false, false},
			"prepareStmt":                              &Entry{true, []any{}, reflect.Bool, false, false},
			"disableNestedTransaction":                 &Entry{false, []any{}, reflect.Bool, false, false},
			"allowGlobalUpdate":                        &Entry{false, []any{}, reflect.Bool, false, false},
			"disableAutomaticPing":                     &Entry{false, []any{}, reflect.Bool, false, false},
			"disableForeignKeyConstraintWhenMigrating": &Entry{false, []any{}, reflect.Bool, false, false},
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
