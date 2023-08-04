package validation

import (
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/config"
	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/walk"
)

const (
	// CurrentElement special key for field name in composite rule sets.
	// Use it if you want to apply rules to the current object element.
	// You cannot apply rules on the root element, these rules will only
	// apply if the rule set is used with composition.
	CurrentElement = ""

	// ExtraRequest extra key used when validating a request so the
	// request's information is accessible to validation rules
	ExtraRequest = "request"
)

// FieldType returned by the GetFieldType function.
const (
	FieldTypeNumeric     = "numeric"
	FieldTypeString      = "string"
	FieldTypeBool        = "bool"
	FieldTypeFile        = "file"
	FieldTypeArray       = "array"
	FieldTypeObject      = "object"
	FieldTypeUnsupported = "unsupported"
)

// ErrorResponse HTTP response format for validation errors.
type ErrorResponse struct {
	Body  *Errors `json:"body,omitempty"`
	Query *Errors `json:"query,omitempty"`
}

// Composable is a partial clone of `goyave.Component`, only
// including the accessors necessary for validation.
// Validators must implement this interface so they
// have access to DB, Config, Language, Logger and ErrLogger.
type Composable interface {
	DB() *gorm.DB
	Config() *config.Config
	Lang() *lang.Language
	Logger() *log.Logger
	ErrLogger() *log.Logger
}

type component struct {
	db        *gorm.DB
	config    *config.Config
	lang      *lang.Language
	logger    *log.Logger
	errLogger *log.Logger
}

// DB get the database instance given through the validation Options.
// Panics if there is none.
func (c *component) DB() *gorm.DB {
	if c.db == nil {
		panic(errors.NewSkip("DB is not set in validation options", 3))
	}
	return c.db
}

// Config get the configuration given through the validation Options.
// Panics if there is none.
func (c *component) Config() *config.Config {
	if c.config == nil {
		panic(errors.NewSkip("Config is not set in validation options", 3))
	}
	return c.config
}

// Lang get the language given through the validation Options.
// Panics if there is none.
func (c *component) Lang() *lang.Language {
	if c.lang == nil {
		panic(errors.NewSkip("Language is not set in validation options", 3))
	}
	return c.lang
}

// Logger get the Logger given through the validation Options.
// Panics if there is none.
func (c *component) Logger() *log.Logger {
	if c.logger == nil {
		panic(errors.NewSkip("Logger is not set in validation options", 3))
	}
	return c.logger
}

// ErrLogger get the Logger given through the validation Options.
// Panics if there is none.
func (c *component) ErrLogger() *log.Logger {
	if c.errLogger == nil {
		panic(errors.NewSkip("ErrLogger is not set in validation options", 3))
	}
	return c.errLogger
}

// Options all the parameters required by `Validate()`.
//
// Only `Data`, `Rules` and `Language` are mandatory. However, it is recommended
// to provide values for all the options in case a `Validator` requires them to function.
type Options struct {
	Data  any
	Rules Ruler

	Now       time.Time
	Extra     map[string]any
	Language  *lang.Language
	DB        *gorm.DB
	Config    *config.Config
	Logger    *log.Logger
	ErrLogger *log.Logger

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
}

// Context is a structure unique per `Validator.Validate()` execution containing
// all the data required by a validator.
type Context struct {
	Data any

	// Extra the map of Extra from the validation Options.
	Extra              map[string]any
	Value              any
	Parent             any
	Field              *Field
	arrayElementErrors []int
	fieldName          string
	Now                time.Time

	// The name of the field under validation
	Name string

	errors []error
	valid  bool // Set to false if there was at least one validation error on the field
}

// AddError adds an error to the validation context. This is NOT supposed
// to be used when the field under validation doesn't match the rule, but rather
// when there has been an operation error (such as a database error).
func (c *Context) AddError(err ...error) {
	for _, e := range err {
		c.errors = append(c.errors, errors.NewSkip(e, 3)) // Skipped: runtime.Callers, NewSkip, this func
	}
}

// AddArrayElementValidationErrors marks a child element to the field currently under validation
// as invalid. This is useful when a validation rule validates an array and wants to
// precisely mark which element in the array is invalid.
func (c *Context) AddArrayElementValidationErrors(index ...int) {
	c.arrayElementErrors = append(c.arrayElementErrors, index...)
}

// Errors returns this validation context's errors.
// The errors returned are NOT validation errors but operation errors (such as database error).
// Because each rule on each field has its own Context, the returned array will only contain
// errors related to the current field and the current rule.
func (c *Context) Errors() []error {
	return c.errors
}

// Valid returns false if at least one validator prior to the current one didn't pass
// on the field under validation.
func (c *Context) Valid() bool {
	return c.valid
}

type validator struct {
	validationErrors *Errors
	options          *Options
	now              time.Time
	errors           []error
}

// Validate the given data using the given `Options`.
// If all validation rules pass and no error occurred, the first returned value will be `nil`.
//
// The second returned value is a slice of error that occurred during validation. These
// errors are not validation errors but error raised when a validator could not be executed correctly.
// For example if a validator using the database generated a DB error.
//
// The `Options.Data` may be modified thanks to type rules.
func Validate(options *Options) (*Errors, []error) {
	validator := &validator{
		options:          options,
		now:              lo.Ternary(options.Now.IsZero(), time.Now(), options.Now),
		errors:           []error{},
		validationErrors: &Errors{},
	}

	rules := options.Rules.AsRules()
	for _, field := range rules {
		if *field.Path.Name == CurrentElement {
			// Validate the root element
			fakeParent := map[string]any{CurrentElement: options.Data}
			validator.validateField(*field.Path.Name, field, fakeParent, nil)
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

func (v *validator) validateField(fieldName string, field *Field, walkData any, parentPath *walk.Path) {
	field.Path.Walk(walkData, func(c *walk.Context) {
		parentObject, parentIsObject := c.Parent.(map[string]any)
		shouldDeleteFromParent := v.shouldDeleteFromParent(field, parentIsObject, c.Value)
		if c.Found == walk.Found {
			if shouldDeleteFromParent {
				delete(parentObject, c.Name)
			}

			if v.shouldConvertSingleValueArray(fieldName) {
				c.Value = v.convertSingleValueArray(field, c.Value)
				parentObject[c.Name] = c.Value
			}
		}

		if v.isAbsent(field, c, v.options.Data) {
			return
		}

		if field.Elements != nil {
			// This is an array, validate its elements first so it can be converted to correct type
			if newValue, ok := makeGenericSlice(c.Value); ok {
				replaceValue(c.Value, c)
				c.Value = newValue
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

		if field.prefixDepth > 0 {
			fullPath := appendPath(parentPath, c.Path, c.Index)
			if rootPath := fullPath.Truncate(field.prefixDepth); rootPath != nil {
				// We can use `First` here because the path contains array indexes
				// so we are sure there will be only one match.
				data = rootPath.First(data).Value
			}
		}

		value := c.Value
		valid := true
		for _, validator := range field.Validators {
			if _, ok := validator.(*NullableValidator); ok {
				if value == nil {
					break
				}
				continue
			}

			ctx := &Context{
				Data:      data,
				Extra:     v.options.Extra,
				Value:     value,
				Parent:    c.Parent,
				Field:     field,
				fieldName: fieldName,
				Now:       v.now,
				Name:      c.Name,
				valid:     valid,
			}
			validator.init(v.options) // TODO document a Rules or RuleSet is not meant to be re-used or used concurrently
			ok := validator.Validate(ctx)
			if len(ctx.errors) > 0 {
				valid = false
				v.errors = append(v.errors, ctx.errors...)
				continue
			}
			if !ok {
				valid = false
				errorPath := field.getErrorPath(parentPath, c)
				message := v.getMessage(ctx, validator)
				if fieldName == CurrentElement {
					v.validationErrors.Add(errorPath, message)
				} else {
					v.validationErrors.Add(&walk.Path{Type: walk.PathTypeObject, Next: errorPath}, message)
				}
				continue
			}

			v.processArrayElementErrors(ctx, parentPath, c, validator)

			value = ctx.Value
		}
		// Value may be modified (converting rule), replace it in the parent element
		if !shouldDeleteFromParent {
			replaceValue(value, c)
		}
	})
}

func (v *validator) shouldDeleteFromParent(field *Field, parentIsObject bool, value any) bool {
	return parentIsObject && !field.IsNullable() && value == nil
}

func (v *validator) shouldConvertSingleValueArray(fieldName string) bool {
	return v.options.ConvertSingleValueArrays && fieldName != CurrentElement && !strings.Contains(fieldName, ".") && !strings.Contains(fieldName, "[]")
}

func (v *validator) convertSingleValueArray(field *Field, value any) any {
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

func (v *validator) isAbsent(field *Field, c *walk.Context, data any) bool {
	if c.Found == walk.ParentNotFound { // TODO document field is skipped if parent not found (make sure to also define rules for parents)
		return true
	}
	requiredCtx := &Context{
		Data:   data,
		Extra:  v.options.Extra,
		Value:  c.Value,
		Parent: c.Parent,
		Field:  field,
		Name:   c.Name,
	}
	return !field.IsRequired(requiredCtx) && !(&RequiredValidator{}).Validate(requiredCtx)
}

func (v *validator) processArrayElementErrors(ctx *Context, parentPath *walk.Path, c *walk.Context, validator Validator) {
	if len(ctx.arrayElementErrors) > 0 {
		errorPath := ctx.Field.getErrorPath(parentPath, c)
		message := v.options.Language.Get(v.getLangEntry(ctx, validator)+".element", v.processPlaceholders(ctx, validator)...)
		for _, index := range ctx.arrayElementErrors {
			i := index
			elementPath := errorPath.Clone()
			elementPath.Type = walk.PathTypeArray
			elementPath.Index = &i
			elementPath.Next = &walk.Path{Type: walk.PathTypeElement}
			if ctx.fieldName == CurrentElement {
				v.validationErrors.Add(elementPath, message)
			} else {
				v.validationErrors.Add(&walk.Path{Type: walk.PathTypeObject, Next: elementPath}, message)
			}
		}
	}
}

func (v *validator) getLangEntry(ctx *Context, validator Validator) string {
	langEntry := "validation.rules." + validator.Name()
	if validator.IsTypeDependent() {
		typeValidator := v.findTypeValidator(ctx.Field.Validators)
		if typeValidator == nil {
			langEntry += "." + GetFieldType(ctx.Value)
		} else {
			typeName := typeValidator.Name()
			switch typeValidator.(type) {
			case *Float32Validator, *Float64Validator,
				*IntValidator, *Int8Validator, *Int16Validator, *Int32Validator, *Int64Validator,
				*UintValidator, *Uint8Validator, *Uint16Validator, *Uint32Validator, *Uint64Validator:
				typeName = FieldTypeNumeric
			}
			langEntry += "." + typeName
		}
	}

	lastParent := ctx.Field.Path.LastParent()
	if lastParent != nil && lastParent.Type == walk.PathTypeArray {
		langEntry += ".element"
	}
	return langEntry
}

func (v *validator) processPlaceholders(ctx *Context, validator Validator) []string {
	return append([]string{":field", translateFieldName(v.options.Language, ctx.fieldName)}, validator.MessagePlaceholders(ctx)...)
}

func (v *validator) getMessage(ctx *Context, validator Validator) string {
	langEntry := v.getLangEntry(ctx, validator)
	return v.options.Language.Get(langEntry, v.processPlaceholders(ctx, validator)...)
}

// findTypeValidator find the expected type of a field for a given array dimension.
func (v *validator) findTypeValidator(validators []Validator) Validator {
	for _, validator := range validators {
		if validator.IsType() {
			return validator
		}
	}

	return nil
}

func replaceValue(value any, c *walk.Context) {
	if c.Found != walk.Found {
		return
	}

	if parentObject, ok := c.Parent.(map[string]any); ok {
		parentObject[c.Name] = value
	} else {
		// Parent is slice
		parent := c.Parent.([]any)
		parent[c.Index] = value
	}
}

func makeGenericSlice(original any) ([]any, bool) {
	if o, ok := original.([]any); ok {
		return o, false
	}
	list := reflect.ValueOf(original)
	length := list.Len()
	newSlice := make([]any, 0, length)
	for i := 0; i < length; i++ {
		newSlice = append(newSlice, list.Index(i).Interface())
	}
	return newSlice, true
}

func appendPath(parentPath, childPath *walk.Path, index int) *walk.Path {
	fullPath := childPath
	if parentPath != nil {
		fullPath = parentPath.Clone()
		tail := fullPath.LastParent()
		if tail != nil {
			tail.Next = childPath
		} else {
			fullPath.Type = walk.PathTypeArray
			fullPath.Index = &index
			fullPath.Next = childPath
		}
	}
	return fullPath
}

// GetFieldType returns the non-technical type of the given "value" interface.
// This is used by validation rules to know if the input data is a candidate
// for validation or not and is especially useful for type-dependent rules.
//   - "numeric" (`lang.FieldTypeNumeric`) if the value is an int, uint or a float
//   - "string" (`lang.FieldTypeString`) if the value is a string
//   - "array" (`lang.FieldTypeArray`) if the value is a slice
//   - "file" (`lang.FieldTypeFile`) if the value is a slice of "fsutil.File"
//   - "unsupported" (`lang.FieldTypeUnsupported`) otherwise
func GetFieldType(value any) string {
	return getFieldType(reflect.ValueOf(value))
}

func getFieldType(value reflect.Value) string {
	kind := value.Kind().String()
	switch {
	case strings.HasPrefix(kind, "int"), strings.HasPrefix(kind, "uint") && kind != "uintptr", strings.HasPrefix(kind, "float"):
		return FieldTypeNumeric
	case kind == "string":
		return FieldTypeString
	case kind == "bool":
		return FieldTypeBool
	case kind == "slice":
		if value.Type().String() == "[]fsutil.File" {
			return FieldTypeFile
		}
		return FieldTypeArray
	default:
		if value.IsValid() {
			if _, ok := value.Interface().(map[string]any); ok {
				return FieldTypeObject
			}
		}
		return FieldTypeUnsupported
	}
}

// GetFieldName returns the localized name of the field identified
// by the given path.
func GetFieldName(lang *lang.Language, path *walk.Path) string {
	return translateFieldName(lang, path.String())
}

func translateFieldName(lang *lang.Language, fieldName string) string {
	if i := strings.LastIndex(fieldName, "."); i != -1 {
		fieldName = fieldName[i+1:]
	}
	fieldName = strings.TrimSuffix(fieldName, "[]")
	entry := "validation.fields." + fieldName
	name := lang.Get(entry)
	if name == entry {
		return fieldName
	}
	return name
}
