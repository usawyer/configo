package defaultValues

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultValues(t *testing.T) {
	type Config struct {
		Host     string            `mapstructure:"host" default:"localhost"`
		Port     int               `mapstructure:"port" default:"8080"`
		Enabled  bool              `mapstructure:"enabled" default:"true"`
		Timeout  float64           `mapstructure:"timeout" default:"30.5"`
		Tags     []string          `mapstructure:"tags" default:"[\"tag1\",\"tag2\"]"`
		Settings map[string]string `mapstructure:"settings" default:"{\"key1\":\"value1\"}"`
	}

	cfg := Config{}
	defaults, err := GetDefaultValues(cfg)
	require.NoError(t, err)

	expected := []DefaultInfo{
		{BindKey: "host", DefaultValue: "localhost"},
		{BindKey: "port", DefaultValue: int64(8080)},
		{BindKey: "enabled", DefaultValue: true},
		{BindKey: "timeout", DefaultValue: 30.5},
		{BindKey: "tags", DefaultValue: []string{"tag1", "tag2"}},
		{BindKey: "settings", DefaultValue: map[string]string{"key1": "value1"}},
	}

	assert.EqualValues(t, expected, defaults)
}

func TestGetDefaultValues_InvalidStruct(t *testing.T) {
	var invalidCfg int
	_, err := GetDefaultValues(invalidCfg)
	assert.Error(t, err, "expected an error for non-struct input")
}

func TestGetDefaultValues_NestedStruct(t *testing.T) {
	type Nested struct {
		Timeout int `mapstructure:"timeout" default:"60"`
	}
	type Config struct {
		Host   string `mapstructure:"host" default:"127.0.0.1"`
		Nested Nested
	}

	cfg := Config{}
	defaults, err := GetDefaultValues(cfg)
	require.NoError(t, err)

	expected := []DefaultInfo{
		{BindKey: "host", DefaultValue: "127.0.0.1"},
		{BindKey: "nested.timeout", DefaultValue: int64(60)},
	}

	assert.EqualValues(t, expected, defaults)
}

func TestGetDefaultValues_NoDefaults(t *testing.T) {
	type Config struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	cfg := Config{}
	defaults, err := GetDefaultValues(cfg)
	require.NoError(t, err)

	assert.Len(t, defaults, 0, "no defaults should be added if not specified")
}
