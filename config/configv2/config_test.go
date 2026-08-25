package configv2

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goyave.dev/goyave/v5/slog"
	"goyave.dev/goyave/v5/util/typeutil"
	v "goyave.dev/goyave/v5/validation"

	_ "embed"
)

// TODO full tests for configv2

//go:embed config.test.json
var embedCfgJSON []byte

type CustomConfig struct {
	Connections   []Connection
	CustomSection CustomSection
	Base
	CustomField int
}

type CustomSection struct {
	A typeutil.Undefined[string] `json:",omitzero"` // TODO document optional config entries need typeutil.Undefined (or alias Optional) with omitzero (depending on the unmarshaler multiple struct tags may be needed)
	B float64
}

func (s CustomConfig) RuleSet() v.RuleSet {
	return v.RuleSet{
		{Path: v.CurrentElement, Rules: s.Base.RuleSet()},
		{Path: "CustomField", Rules: v.List{v.Required(), v.Int(), v.Min(2)}},
		{Path: "CustomSection", Rules: v.List{v.Required(), v.Object()}},
		{Path: "CustomSection.A", Rules: v.List{v.Required(), v.String()}},
		{Path: "CustomSection.B", Rules: v.List{v.Required(), v.Float64()}},
		{Path: "Connections", Rules: v.List{v.Required(), v.Array()}},
		{Path: "Connections[]", Rules: Connection{}.RuleSet()},
	}
}

func (s CustomConfig) Default() CustomConfig {
	return CustomConfig{
		Base:        s.Base.Default(),
		CustomField: 1,
		CustomSection: CustomSection{
			// A: "default A", // A doesn't have a default value so should be undefined
			B: 1.2,
		},
		Connections: []Connection{Connection{}.Default()},
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

func TestLoad(t *testing.T) {
	_ = os.Setenv("ENV_TEST", "3")
	cfg, validationErrors, err := Load[CustomConfig](t.Context(), FromBytes(embedCfgJSON, UnsmarshalJSON()))
	logger := slog.New(slog.NewHandler(true, t.Output()))
	require.NoError(t, err)
	if !assert.Nil(t, validationErrors) {
		logger.Error(fmt.Errorf("configuration validation errors"), "errors", validationErrors)
	}

	want := &CustomConfig{
		App: App{
			Name:            "goyave",
			Environment:     "test_json",
			DefaultLanguage: "fr-FR",
			Debug:           false,
		},
		CustomSection: CustomSection{
			A: typeutil.NewUndefined(""),
			B: 6.999,
		},
		CustomField: 13,
		Connections: []Connection{
			{
				Driver: "postgres",
				Host:   "127.0.0.1",
				Port:   65534,
			},
			{
				Driver: "mysql",
				Host:   "127.0.0.1",
				Port:   65535,
			},
		},
	}
	require.Equal(t, want, cfg)

	t.Log(cfg.App.Name)
	t.Log(cfg.App.Environment)
	t.Log(cfg.App.DefaultLanguage)
	t.Log(cfg.App.Debug)
	t.Log(cfg.CustomField)
	t.Log(cfg.CustomSection)
}
