package env

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnvStructure_Simple(t *testing.T) {
	type simpleConfig struct {
		Host    string `env:"HOST" default:"localhost" help:"Hostname"`
		Port    int    `mapstructure:"port" default:"8080" help:"Port number"`
		ignored bool   // unexported, should be ignored
		Skip    string `env:"-" default:"shouldNotAppear" help:"Should be ignored"`
		NoTags  string // no tags, should default to uppercase field name: "NOTAGS"
	}

	var lines []EnvInfo
	parseEnvStructure(reflect.TypeOf(simpleConfig{}), "", "", &lines)

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	tests := map[string]struct {
		DefaultValue string
		HelpText     string
	}{
		"HOST": {
			DefaultValue: "localhost",
			HelpText:     "Hostname",
		},
		"PORT": {
			DefaultValue: "8080",
			HelpText:     "Port number",
		},
		"NOTAGS": {
			DefaultValue: "", // no default
			HelpText:     "", // no help
		},
	}

	for _, info := range lines {
		expected, found := tests[info.EnvVar]
		if !found {
			t.Errorf("unexpected EnvVar: %s", info.EnvVar)
			continue
		}
		if info.DefaultValue != expected.DefaultValue {
			t.Errorf("EnvVar %s default value = %q, want %q",
				info.EnvVar, info.DefaultValue, expected.DefaultValue)
		}
		if info.HelpText != expected.HelpText {
			t.Errorf("EnvVar %s help text = %q, want %q",
				info.EnvVar, info.HelpText, expected.HelpText)
		}
	}
}

func TestParseEnvStructure_NestedStruct(t *testing.T) {
	type nestedStruct struct {
		NestedKey string `env:"NESTED_KEY" default:"nested" help:"Nested key"`
	}

	type configWithNested struct {
		MainKey   string       `env:"MAIN_KEY" default:"main" help:"Main key"`
		Nested    nestedStruct `help:"Should parse nested struct fields"`
		IgnoredNS nestedStruct `env:"-" help:"Should not parse nested fields"`
	}

	lines := GetEnvs(configWithNested{})

	// We expect "MAIN_KEY" and "NESTED_KEY" but not "IGNOREDNS"
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestGetEnvs(t *testing.T) {
	type Config struct {
		Host    string  `env:"HOST" default:"localhost" help:"Database host"`
		Port    int     `env:"PORT" default:"5432" help:"Database port"`
		Enabled bool    `env:"ENABLED" default:"true" help:"Enable feature"`
		Timeout float64 `env:"TIMEOUT" default:"30.5" help:"Request timeout"`
	}

	cfg := Config{}
	envs := GetEnvs(cfg)
	require.NotEmpty(t, envs)

	expected := []EnvInfo{
		{EnvVar: "HOST", DefaultValue: "localhost", HelpText: "Database host", BindKey: "host", ValueType: "string"},
		{EnvVar: "PORT", DefaultValue: "5432", HelpText: "Database port", BindKey: "port", ValueType: "int"},
		{EnvVar: "ENABLED", DefaultValue: "true", HelpText: "Enable feature", BindKey: "enabled", ValueType: "bool"},
		{EnvVar: "TIMEOUT", DefaultValue: "30.5", HelpText: "Request timeout", BindKey: "timeout", ValueType: "float64"},
	}

	assert.EqualValues(t, expected, envs)
}

func TestGetEnvs_NoEnvTags(t *testing.T) {
	type Config struct {
		Host string `mapstructure:"dbhost"`
		Port int
	}

	cfg := Config{}
	envs := GetEnvs(cfg)

	expected := []EnvInfo{
		{EnvVar: "DBHOST", DefaultValue: "", HelpText: "", BindKey: "dbhost", ValueType: "string"},
		{EnvVar: "PORT", DefaultValue: "", HelpText: "", BindKey: "port", ValueType: "int"},
	}

	assert.EqualValues(t, expected, envs)
}

func TestGetEnvs_EnvTagOverride(t *testing.T) {
	type Config struct {
		Host string `env:"CUSTOM_HOST" mapstructure:"host" default:"localhost" help:"Custom host"`
	}

	cfg := Config{}
	envs := GetEnvs(cfg)

	expected := []EnvInfo{
		{EnvVar: "CUSTOM_HOST", DefaultValue: "localhost", HelpText: "Custom host", BindKey: "host", ValueType: "string"},
	}

	assert.EqualValues(t, expected, envs)
}

func TestGetEnvs_StructWithEnvDash(t *testing.T) {
	type Nested struct {
		Timeout int `env:"TIMEOUT" default:"60" help:"Nested timeout"`
	}
	type Config struct {
		Host   string `env:"HOST" default:"127.0.0.1" help:"Main host"`
		Nested Nested `env:"-"`
	}

	cfg := Config{}
	envs := GetEnvs(cfg)

	expected := []EnvInfo{
		{EnvVar: "HOST", DefaultValue: "127.0.0.1", HelpText: "Main host", BindKey: "host", ValueType: "string"},
	}

	assert.EqualValues(t, expected, envs)
}

func TestGetEnvs_FieldWithEnvDash(t *testing.T) {
	type Config struct {
		Host string `env:"-" mapstructure:"host" default:"localhost" help:"Should be ignored"`
		Port int    `env:"PORT" default:"5432" help:"Database port"`
	}

	cfg := Config{}
	envs := GetEnvs(cfg)

	expected := []EnvInfo{
		{EnvVar: "PORT", DefaultValue: "5432", HelpText: "Database port", BindKey: "port", ValueType: "int"},
	}

	assert.EqualValues(t, expected, envs)
}

func TestGetEnvs_InvalidType(t *testing.T) {
	var invalidCfg int
	envs := GetEnvs(invalidCfg)
	assert.Len(t, envs, 0, "No env variables should be parsed from non-struct types")
}
