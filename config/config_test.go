package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goyave.dev/goyave/v5/util/typeutil"
	v "goyave.dev/goyave/v5/validation"

	_ "embed"
)

// TODO full tests for configv2

//go:embed config.test.json
var embedCfgJSON []byte

type CustomConfig struct {
	CustomConnections []Connection
	CustomSection     CustomSection
	Base              // Proper composition with custom fields correctly supported for config extension.
	CustomField       string
	Slice             []string
	TwoDimSlice       [][]string
}

type CustomSection struct {
	A      typeutil.Undefined[string] `json:",omitzero"` // Just for the sake of testing that Undefined is properly supported.
	B      float64
	Number int64 // Number input from an env variable, checking that conversion works as expected.
}

func (s CustomConfig) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: s.Base.RuleSet()},
		{Path: "CustomField", Rules: v.List{v.Required(), v.String(), v.Max(50)}},
		{Path: "CustomSection", Rules: v.List{v.Required(), v.Object()}},
		{Path: "CustomSection.A", Rules: v.List{v.String()}},
		{Path: "CustomSection.B", Rules: v.List{v.Required(), v.Float64()}},
		{Path: "CustomSection.Number", Rules: v.List{v.Required(), v.Int64(), v.Min(2)}},
		{Path: "CustomConnections", Rules: v.List{v.Required(), v.Array()}},
		{Path: "CustomConnections[]", Rules: Connection{}.RuleSet()},
		{Path: "Slice", Rules: v.List{v.Required(), v.Array()}},
		{Path: "Slice[]", Rules: v.List{v.String()}},
		{Path: "TwoDimSlice", Rules: v.List{v.Required(), v.Array()}},
		{Path: "TwoDimSlice[]", Rules: v.List{v.Array()}},
		{Path: "TwoDimSlice[][]", Rules: v.List{v.String()}},
	}
}

func (s CustomConfig) Default() CustomConfig {
	return CustomConfig{
		Base:        s.Base.Default(),
		CustomField: "custom default",
		CustomSection: CustomSection{
			// A: "default A", // A doesn't have a default value so should be undefined
			B:      1.2,
			Number: -1,
		},
		CustomConnections: []Connection{Connection{}.Default()},
		Slice:             []string{},
		TwoDimSlice: [][]string{
			{},
			{"default value"},
		},
	}
}

type Connection struct {
	Driver string
	Host   string
	Port   int
}

func (s Connection) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: v.List{v.Object()}},
		{Path: "Driver", Rules: v.List{v.Required(), v.String(), v.Min(1)}},
		{Path: "Host", Rules: v.List{v.Required(), v.String(), v.Min(1)}},
		{Path: "Port", Rules: v.List{v.Required(), v.Int(), v.Between(0, 65535)}},
	}
}

func (s Connection) Default() Connection {
	return Connection{
		Driver: "",
		Host:   "0.0.0.0",
		Port:   5432,
	}
}

type wrongType int

func (wrongType) RuleSet() v.RuleSet {
	return nil
}

func TestLoad(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		t.Setenv("TEST_VAR", "3")
		t.Setenv("TEST_VAR_2", "4")
		t.Setenv("TEST_VAR_3", "5")
		t.Setenv("TEST_VAR_4", "6")
		t.Setenv("TEST_VAR_5", "7")
		t.Setenv("NUMBER", "5")
		cfg, err := Load[CustomConfig](t.Context(), FromBytes(embedCfgJSON, UnsmarshalJSON()))
		require.NoError(t, err)

		want := &CustomConfig{
			App: App{
				Name:            "goyave", // Default values are set
				Environment:     "test_json",
				DefaultLanguage: "fr-FR", // Default values are overridden
				Debug:           false,
				// "Nested" key ignored because it's not in the target struct
			},
			Server: Server{}.Default(), // Server not in the config file at all so expect defaults
			CustomSection: CustomSection{
				A:      typeutil.NewUndefined(""), // Empty string provided in the config file so the value is "present"
				B:      6.999,
				Number: 5, // Interpolated from "NUMBER" env var
			},
			CustomField: "interpolated:3 | multiple: 3/4", // Interpolated from TEST_VAR and TEST_VAR_2
			CustomConnections: []Connection{
				{
					Driver: "postgres",  // First element is a default value.
					Host:   "127.0.0.1", // It was properly overridden
					Port:   65534,
				},
				{
					Driver: "mysql", // Second element was not in defaults, it was appended.
					Host:   "127.0.0.1",
					Port:   65535,
				},
			},
			Slice: []string{"value 1", "5"}, // Second element has env var interpolation TEST_VAR_3
			TwoDimSlice: [][]string{
				{}, // Default empty slice left empty
				{
					"6", // Default value overridden, env var interpolation TEST_VAR_4
					"7", // Env var interpolation TEST_VAR_5
				},
				{"appended"},
			},
		}
		require.Equal(t, want, cfg)
	})

	t.Run("validation", func(t *testing.T) {
		// Default values don't pass validation on purpose in this test
		// Check validation fails after interpolation too
		t.Setenv("TEST_VAR", "not a number")
		cfgJSON := `{
			"CustomSection": {
				"Number": "${TEST_VAR}"
			}	
		}`
		src := FromBytes([]byte(cfgJSON), UnsmarshalJSON())

		cfg, err := Load[CustomConfig](t.Context(), src)
		assert.Nil(t, cfg)
		require.Error(t, err)

		validationErrs, ok := errors.AsType[*v.Errors](err)
		require.True(t, ok)

		want := &v.Errors{
			Fields: v.FieldsErrors{
				"CustomConnections": &v.Errors{
					Elements: v.ArrayErrors{
						0: &v.Errors{
							Fields: v.FieldsErrors{
								"Driver": &v.Errors{
									Errors: []string{"The Driver must be at least 1 characters."},
								},
							},
						},
					},
				},
				"CustomSection": &v.Errors{
					Fields: v.FieldsErrors{
						"Number": &v.Errors{
							Errors: []string{"The Number must be an integer."},
						},
					},
				},
			},
		}

		assert.Equal(t, want, validationErrs)
	})

	t.Run("load_many_merge", func(t *testing.T) {
		cfg1 := `{"App": {"Name": "test_app"}}`
		cfg2 := `{"App": {"Debug": true}}` // Should have no effect, it's the default value
		cfg3 := `{"CustomSection": {"Number": 2}}`
		cfg4 := `{"CustomConnections": [{"Driver": "postgres"}]}` // The first CustomConnections element will be merged

		src1 := FromBytes([]byte(cfg1), UnsmarshalJSON())
		src2 := FromBytes([]byte(cfg2), UnsmarshalJSON())
		src3 := FromBytes([]byte(cfg3), UnsmarshalJSON())
		src4 := FromBytes([]byte(cfg4), UnsmarshalJSON())
		cfg, err := Load[CustomConfig](t.Context(), src1, src2, src3, src4)
		require.NoError(t, err)

		want := new(CustomConfig{}.Default())
		want.App.Name = "test_app"
		want.CustomSection.Number = 2
		want.CustomConnections[0].Driver = "postgres"
		require.Equal(t, want, cfg)
		assert.False(t, cfg.CustomSection.A.Present) // Extra test for testutil.Undefined support
	})

	t.Run("load_default_source", func(t *testing.T) {
		t.Run("ENOENT", func(t *testing.T) {
			cfg, err := Load[CustomConfig](t.Context()) // Should use Default() = config.json
			assert.Nil(t, cfg)
			require.ErrorContains(t, err, "no such file")
		})

		t.Run("OK", func(t *testing.T) {
			t.Setenv("ENV", "test")
			t.Setenv("TEST_VAR", "3")
			t.Setenv("TEST_VAR_2", "4")
			t.Setenv("NUMBER", "5")

			_, err := Load[CustomConfig](t.Context()) // Should use Default() = config.test.json
			require.NoError(t, err)
		})
	})

	t.Run("non_struct_type", func(t *testing.T) {
		cfg, err := Load[wrongType](t.Context())
		assert.Nil(t, cfg)
		assert.ErrorContains(t, err, "T must be a structure")
	})

	t.Run("LoadDefault", func(t *testing.T) {
		assert.Equal(t, new(Base{}.Default()), LoadDefault())
	})
}
