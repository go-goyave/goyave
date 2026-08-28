package config

import (
	"context"
	"os"

	"goyave.dev/goyave/v5/lang"
	"goyave.dev/goyave/v5/util/errors"
	"goyave.dev/goyave/v5/util/typeutil"
	"goyave.dev/goyave/v5/validation"
)

// Section represents a configuration section like "App" or "Database".
// All structures used in configuration must implement this interface so
// their content can be validated when loaded.
//
// The main configuration structure is also considered a root `Section`.
//
// All sections should be validated as objects.
//
// If the Section contains sub-sections, it should generate its validation RuleSet
// using the sub-sections `RuleSet()` implementation so nested sections are also
// correctly validated.
//
//	type CustomConfig struct {
//		config.Base
//		ArraySection   []ArraySection
//		CustomSection CustomSection
//		CustomField   int
//	}
//
//	func (s CustomConfig) RuleSet() v.RuleSet {
//		return v.RuleSet{
//			{Path: v.CurrentElement, Rules: s.Base.RuleSet()},
//			{Path: "CustomField", Rules: v.List{v.Required(), v.Int(), v.Min(2)}},
//			{Path: "CustomSection", Rules: s.CustomSection.RuleSet()},
//			{Path: "ArraySection", Rules: v.List{v.Required(), v.Array()}},
//			{Path: "ArraySection[]", Rules: ArraySection{}.RuleSet()},
//		}
//	}
//
//	type CustomSection struct {
//		A string
//		B float64
//	}
//
//	func (CustomSection) RuleSet() v.RuleSet {
//		return v.RuleSet{
//			{Path: v.CurrentElement, Rules: v.List{v.Required(), v.Object()}},
//			{Path: "A", Rules: v.List{v.Required(), v.String(), v.Min(1)}},
//			{Path: "B", Rules: v.List{v.Required(), v.Float64()}},
//		}
//	}
//
//	type ArraySection struct {
//		C string
//		D int8
//	}
//
//	func (ArraySection) RuleSet() v.RuleSet {
//		return v.RuleSet{
//			{Path: v.CurrentElement, Rules: v.List{v.Object()}},
//			{Path: "C", Rules: v.List{v.Required(), v.String(), v.Min(1)}},
//			{Path: "D", Rules: v.List{v.Required(), v.Int8()}},
//		}
//	}
type Section interface {
	RuleSet() validation.RuleSet
}

// DefaultValuer can be optionally implemented by `Section` implementations
// to provide default configuration values.
// Type T should be equal to the implementation's own type.
//
// If the Section contains sub-sections, it should call the sub-sections' `Default()`
// method as well in order to correctly populate the nested default values.
//
//	type MyConfig struct {
//		Value      string
//		SubSection SubSection
//	}
//
//	func (s MyConfig) Default() MyConfig {
//		return MyConfig{
//			Value:      "default value",
//			SubSection: s.SubSection.Default(),
//		}
//	}
//
//	type SubSection struct {
//		SubValue string
//	}
//
//	func (SubSection) Default() SubSection {
//		return SubSection{
//			SubValue: "default subvalue",
//		}
//	}
type DefaultValuer[T Section] interface {
	Default() T
}

// Load and validate configuration from one or multiple sources.
//
//	//go:embed config.json
//	var embedCfgJSON []byte
//
//	logger := slog.New(slog.NewHandler(true, os.Stderr))
//	cfg, validationErrors, err := Load[CustomConfig](context.Background(), config.FromBytes(embedCfgJSON, config.UnsmarshalJSON()))
//
//	if err != nil {
//		logger.Error(err)
//		return
//	}
//
//	if validationErrors != nil {
//		logger.Error(fmt.Errorf("configuration validation errors"), "errors", validationErrors)
//		return
//	}
//
// Type T shouldn't be a pointer.
//
//   - First, an empty struct of type T is initialized. If it implements `DefaultValuer[T]`,
//     the default values are loaded from there.
//   - All given sources are read sequentially and merged into the default values.
//   - The result is validated using T's rule set and converted to the expected struct type `*T`.
//
// Environment variables can be interpolated: occurrences of `${ENV_VAR}` or `$ENV_VAR` are automatically
// replaced with the value of the corresponding environment variables. References to undefined
// variables are replaced by the empty string.
//
// When merging multiple sources together (or a single source into the default values), only the keys
// that are defined in the merged source are overridden.
//
// For example, if the source contains only the "App.Name" entry, all default values will be kept and the
// "App.Name" entry will be overridden with the value from the source.
//
// When merging arrays/slices, each element is updated in-place recursively. Extra elements are appended.
// E.g.: the current configuration contains only one object element and the merged configuration contains two.
// The first element will be updated with the object keys/values present in the merged configuration. The second
// element will be appended as-is.
//
//	Current configuration:
//	{
//		"Slice": [
//			{"Name": "foo", "Value": 1}
//		]
//	}
//
//	Merged configuration:
//	{
//		"Slice": [
//			{"Value": 2}
//			{"Name": "bar", "Value": 3}
//		]
//	}
//
//	Result:
//	{
//		"Slice": [
//			{"Name": "foo", "Value": 2}
//			{"Name": "bar", "Value": 3}
//		]
//	}
//
// Validation occurs only once on the result of all merges.
func Load[T Section](ctx context.Context, sources ...Source) (*T, *validation.Errors, error) {
	if len(sources) == 0 {
		sources = []Source{Default()}
	}

	var defaultCfg T
	if d, ok := any(defaultCfg).(DefaultValuer[T]); ok {
		defaultCfg = d.Default()
	}
	cfg, err := typeutil.Convert[map[string]any](defaultCfg)
	if err != nil {
		return nil, nil, errors.New([]error{errors.New("failed to convert default config to map; T must be a structure"), errors.New(err)})
	}

	for _, source := range sources {
		raw, err := source.Read()
		if err != nil {
			return nil, nil, errors.New([]error{errors.New("failed to unmarshal config"), errors.New(err)})
		}
		mapMerge(cfg, raw)
	}

	opt := &validation.Options{
		Context:  ctx,
		Data:     cfg,
		Rules:    defaultCfg.RuleSet(),
		Language: lang.Default,
	}
	errsBag, errs := validation.Validate(opt)
	if errs != nil {
		return nil, nil, errors.New(append([]error{errors.Errorf("failed to validate config")}, errs...))
	}

	if errsBag != nil {
		return nil, errsBag, nil
	}

	loaded, err := typeutil.Convert[*T](cfg)
	if err != nil {
		return nil, nil, errors.Errorf("failed to convert config map to struct: %w", err)
	}

	return loaded, nil, nil
}

// LoadDefault returns the default values for the [Base] config.
// Configuration files are not loaded by calling this function.
func LoadDefault() *Base {
	cfg := Base{}.Default()
	return &cfg
}

// mapMerge recursively merges src into dst.
// If src isn't a map, returns immediately.
// This preserves explicitly-provided zero values that would be otherwise
// mishandled with typeutil.Copy.
func mapMerge(dst map[string]any, src any) any {
	srcMap, ok := src.(map[string]any)
	if !ok {
		return expandEnv(src)
	}
	for k, v := range srcMap {
		switch dstVal := dst[k].(type) {
		case map[string]any:
			dst[k] = mapMerge(dstVal, v)
		case []any:
			dst[k] = sliceMerge(dstVal, v)
		default:
			dst[k] = expandEnv(v)
		}
	}
	return dst
}

func sliceMerge(dst []any, src any) any {
	srcSlice, ok := src.([]any)
	if !ok {
		return expandEnv(src)
	}
	for i, v := range srcSlice {
		if i >= len(dst) {
			dst = append(dst, v)
			continue
		}
		switch dstVal := dst[i].(type) {
		case map[string]any:
			dst[i] = mapMerge(dstVal, v)
		case []any:
			dst[i] = sliceMerge(dstVal, v)
		default:
			dst[i] = expandEnv(v)
		}
	}
	return dst
}

func expandEnv(v any) any {
	str, ok := v.(string)
	if !ok {
		return v
	}

	// TODO Overriding using command-line flags?
	// e.g.: -config App.Name="override"
	// -> Possible to implement later without breaking changes (track the Path in mapMerge/sliceMerge)
	// Limitation: walk.Path doesn't support parsing array indexes.
	// For now I don't really see the point since we can use env variables
	return os.ExpandEnv(str) // TODO document that "environment variable is not set" error doesn't exist anymore
}
