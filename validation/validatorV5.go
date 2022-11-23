package validation

import (
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"goyave.dev/goyave/v4/config"
	"goyave.dev/goyave/v4/lang"
	"goyave.dev/goyave/v4/util/walk"
)

type Options struct {
	Data  any
	Rules RulerV5

	// ConvertSingleValueArrays set to true to convert fields that are expected
	// to be an array into an array with a single value.
	//
	// It is recommended to set this option to `true` when validating url-encoded requests.
	// For example, if set to `false`:
	//  field=A         --> map[string]any{"field": "A"}
	//  field=A&field=B --> map[string]any{"field": []string{"A", "B"}}
	// If set to `true` and `field` has the `Array` rule:
	//  field=A         --> map[string]any{"field": []string{"A"}}
	//  field=A&field=B --> map[string]any{"field": []string{"A", "B"}}
	ConvertSingleValueArrays bool
	Language                 *lang.Language
	DB                       *gorm.DB
	Config                   *config.Config
	Extra                    map[string]any
}

type ContextV5 struct {
	Options *Options
	Data    any
	Extra   map[string]any
	Value   any
	Parent  any
	Field   *FieldV5
	Now     time.Time

	// The name of the field under validation
	Name string

	errors []error
}

// DB get the database instance given through the validation Options.
// Panics if there is none.
func (c *ContextV5) DB() *gorm.DB {
	if c.Options.DB == nil {
		panic("DB is not set in validation options")
	}
	return c.Options.DB
}

// Config get the configuration given through the validation Options.
// Panics if there is none.
func (c *ContextV5) Config() *config.Config {
	if c.Options.Config == nil {
		panic("Config is not set in validation options")
	}
	return c.Options.Config
}

// AddError adds an error to the validation context. This is NOT supposed
// to be used when the field under validation doesn't match the rule, but rather
// when there has been an operation error (such as a database error).
func (c *ContextV5) AddError(err ...error) {
	c.errors = append(c.errors, err...)
}

// Errors returns this validation context's errors. Because each rule on each field
// has its own Context, the returned array will only contain errors related to the
// current field and the current rule.
func (c *ContextV5) Errors() []error {
	return c.errors
}

type validator struct {
	validationErrors *ErrorsV5
	options          *Options
	now              time.Time
	errors           []error
}

func ValidateV5(options *Options) (*ErrorsV5, []error) {
	validator := &validator{
		options:          options,
		now:              time.Now(),
		errors:           []error{},
		validationErrors: &ErrorsV5{},
	}

	rules := options.Rules.AsRules()
	for _, field := range rules {
		if *field.Path.Name == CurrentElement {
			// Validate the root element
			fakeParent := map[string]any{CurrentElement: options.Data}
			validator.validateField(*field.Path.Name, field, fakeParent, nil) // TODO field name for the root element?
			options.Data = fakeParent[CurrentElement]
		} else {
			validator.validateField(*field.Path.Tail().Name, field, options.Data, nil)
		}
	}

	if len(validator.errors) != 0 {
		return nil, validator.errors
	}
	if len(validator.validationErrors.Errors) != 0 || len(validator.validationErrors.Elements) != 0 || len(validator.validationErrors.Fields) != 0 {
		return validator.validationErrors, nil
	}
	return nil, nil
}

func (v *validator) validateField(fieldName string, field *FieldV5, walkData any, parentPath *walk.Path) {
	field.Path.Walk(walkData, func(c *walk.Context) {
		parentObject, parentIsObject := c.Parent.(map[string]any)
		if c.Found == walk.Found {
			if parentIsObject && !field.IsNullable() && c.Value == nil {
				delete(parentObject, c.Name)
			}
			// TODO if the parent is an array, should be removed too!

			if v.shouldConvertSingleValueArray(fieldName) {
				c.Value = v.convertSingleValueArray(field, c.Value, parentObject)
				parentObject[c.Name] = c.Value
			}
		}

		if v.isAbsent(field, c, v.options.Data) {
			return
		}

		if field.Elements != nil {
			// This is an array, validate its elements first so it can be converted to correct type
			if _, ok := c.Value.([]any); !ok {
				if newValue, ok := makeGenericSlice(c.Value); ok {
					replaceValue(c.Value, c)
					c.Value = newValue
				}
			}

			path := c.Path
			if parentPath != nil {
				clone := parentPath.Clone()
				tail := clone.Tail()
				tail.Type = walk.PathTypeArray
				tail.Index = &c.Index
				tail.Next = path.Next
				path = clone
			}
			v.validateField(fieldName+"[]", field.Elements, c.Value, path)
		}

		data := v.options.Data
		if rootPath := c.Path.Truncate(field.prefixDepth); rootPath != nil {
			rootPath.Walk(walkData, func(ctx *walk.Context) {
				// TODO use First function
				// This function will be called only once because
				// the path contains indexes.
				data = ctx.Value
			})
		}

		value := c.Value
		for _, validator := range field.Validators {
			if _, ok := validator.(*NullableValidator); ok {
				if value == nil {
					break
				}
				continue
			}

			ctx := &ContextV5{
				Options: v.options,
				Data:    data,
				Extra:   v.options.Extra, // TODO document extra are not scoped anymore (and not copied)
				Value:   value,
				Parent:  c.Parent,
				Field:   field,
				Now:     v.now,
				Name:    c.Name,
			}
			ok := validator.Validate(ctx)
			if len(ctx.errors) > 0 {
				v.errors = append(v.errors, ctx.errors...)
				continue
			}
			if !ok {
				errorPath := field.getErrorPath(parentPath, c)
				message := v.getMessage(fieldName, field, validator, reflect.ValueOf(value))
				if fieldName == CurrentElement {
					v.validationErrors.Add(errorPath, message)
				} else {
					v.validationErrors.Add(&walk.Path{Type: walk.PathTypeObject, Next: errorPath}, message)
				}
				continue
			}

			value = ctx.Value
		}
		// Value may be modified (converting rule), replace it in the parent element
		replaceValue(value, c)
	})
}

func (v *validator) shouldConvertSingleValueArray(fieldName string) bool {
	return v.options.ConvertSingleValueArrays && fieldName != CurrentElement && !strings.Contains(fieldName, ".") && !strings.Contains(fieldName, "[]")
}

func (v *validator) convertSingleValueArray(field *FieldV5, value any, data map[string]any) any {
	rv := reflect.ValueOf(value)
	kind := rv.Kind().String()
	if field.IsArray() && kind != "slice" {
		rt := reflect.TypeOf(value)
		slice := reflect.MakeSlice(reflect.SliceOf(rt), 0, 1)
		slice = reflect.Append(slice, rv)
		return slice.Interface()
	}
	return value
}

func (v *validator) isAbsent(field *FieldV5, c *walk.Context, data any) bool {
	requiredCtx := &ContextV5{
		Options: v.options,
		Data:    data,
		Extra:   v.options.Extra,
		Value:   c.Value,
		Parent:  c.Parent,
		Field:   field,
		Name:    c.Name,
	}
	return !field.IsRequired(requiredCtx) && !(&RequiredValidator{}).Validate(requiredCtx)
}

func (v *validator) getMessage(fieldName string, field *FieldV5, validator Validator, value reflect.Value) string {
	langEntry := "validation.rules." + validator.Name()
	if validator.IsTypeDependent() {
		expectedType := v.findTypeRule(field.Validators)
		if expectedType == FieldTypeUnsupported {
			langEntry += "." + getFieldType(value)
		} else {
			if strings.HasPrefix(expectedType, "int") || strings.HasPrefix(expectedType, "uint") || strings.HasPrefix(expectedType, "float") {
				expectedType = FieldTypeNumeric
			}
			langEntry += "." + expectedType
		}
	}

	lastParent := field.Path.LastParent()
	if lastParent != nil && lastParent.Type == walk.PathTypeArray {
		langEntry += ".array"
	}

	return v.options.Language.Get(langEntry, append([]string{":field", translateFieldName(v.options.Language, fieldName)}, validator.MessagePlaceholders(v.options.Language)...)...)
}

// findTypeRule find the expected type of a field for a given array dimension.
func (v *validator) findTypeRule(rules []Validator) string {
	for _, rule := range rules {
		if rule.IsType() {
			return rule.Name()
		}
	}

	return FieldTypeUnsupported
}
