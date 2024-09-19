package configManager

import (
	"os"
	"reflect"
	"slices"
	"testing"
)

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"  default:"defaulthost"`
	Port int    `mapstructure:"port" default:"8081"`
}

type TestConfig struct {
	AppName  string         `mapstructure:"appName" env:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Statuses []string       `mapstructure:"statuses" desc:"Статусы" default:"a,b,c,aa,ab"`
	Enable   bool           `mapstructure:"enable" desc:"Флаг для включения определенной функции" default:"true"`
}

func (c TestConfig) Validate() error {
	//if c.Database.URL == "" {
	//	return fmt.Errorf("Database URL cannot be empty")
	//}
	//if c.Server.Port <= 0 {
	//	return fmt.Errorf("Server port must be greater than zero")
	//}
	return nil
}

func createTempYAMLConfig(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpFile.Name()
}
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Failed to set env variable %s: %v", key, err)
	}
}
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Failed to unset env variable %s: %v", key, err)
	}
}

// Тестирование загрузки конфигурации из YAML-файла
func TestConfigManager_LoadNestedConfigFromYAML(t *testing.T) {
	yamlContent := `
appName: "testapp"
database:
  url: "postgres://localhost:5432/db"
  username: "dbuser"
  password: "dbpass"
server:
  host: "localhost"
  port: 8080
statuses:
  - "a"
  - "b"
  - "c"
Enable: false
`
	configPath := createTempYAMLConfig(t, yamlContent)
	defer os.Remove(configPath)

	settings := Settings{
		ConfigFilePath: configPath,
	}

	cm, err := NewConfigManager[TestConfig](settings)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm.Config()

	// Проверяем значения
	if config.AppName != "testapp" {
		t.Errorf("Expected AppName to be 'testapp', got '%s'", config.AppName)
	}
	if config.Database.URL != "postgres://localhost:5432/db" {
		t.Errorf("Expected Database.URL to be 'postgres://localhost:5432/db', got '%s'", config.Database.URL)
	}
	if config.Database.Username != "dbuser" {
		t.Errorf("Expected Database.Username to be 'dbuser', got '%s'", config.Database.Username)
	}
	if config.Database.Password != "dbpass" {
		t.Errorf("Expected Database.Password to be 'dbpass', got '%s'", config.Database.Password)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected Server.Host to be 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected Server.Port to be 8080, got %d", config.Server.Port)
	}
	expectedStatuses := []string{"a", "b", "c"}
	if !slices.Equal(config.Statuses, expectedStatuses) {
		t.Errorf("Expected Statuses to be %v, got %v", expectedStatuses, config.Statuses)
	}
	if config.Enable != false {
		t.Errorf("Expected Enable to be false, got %v", config.Enable)
	}
}

// Тестирование переопределения значений через переменные окружения
// Важно: По умолчанию viper преобразует имена переменных окружения в нижний регистр и использует символ подчеркивания _ в качестве разделителя.
func TestConfigManager_NestedEnvOverridesYAML(t *testing.T) {
	yamlContent := `
appName: "testapp"
database:
  url: "postgres://localhost:5432/db"
  username: "dbuser"
  password: "dbpass"
server:
  host: "localhost"
  port: 8080
`
	configPath := createTempYAMLConfig(t, yamlContent)
	defer os.Remove(configPath)

	// Устанавливаем переменные окружения для вложенных полей
	setEnv(t, "APP", "envapp")
	setEnv(t, "DATABASE_URL", "postgres://env-db:5432/db")
	setEnv(t, "DATABASE_USERNAME", "envuser")
	setEnv(t, "SERVER_PORT", "9090")
	setEnv(t, "STATUSES", "a,b")
	setEnv(t, "ENABLE", "true")
	defer unsetEnv(t, "APP")
	defer unsetEnv(t, "DATABASE_URL")
	defer unsetEnv(t, "DATABASE_USERNAME")
	defer unsetEnv(t, "SERVER_PORT")
	defer unsetEnv(t, "STATUSES")
	defer unsetEnv(t, "ENABLE")

	settings := Settings{
		ConfigFilePath: configPath,
	}

	cm, err := NewConfigManager[TestConfig](settings)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm.Config()

	// Проверяем, что значения из переменных окружения переопределили YAML
	if config.AppName != "envapp" {
		t.Errorf("Expected AppName to be 'envapp', got '%s'", config.AppName)
	}
	if config.Database.URL != "postgres://env-db:5432/db" {
		t.Errorf("Expected Database.URL to be 'postgres://env-db:5432/db', got '%s'", config.Database.URL)
	}
	if config.Database.Username != "envuser" {
		t.Errorf("Expected Database.Username to be 'envuser', got '%s'", config.Database.Username)
	}
	// Пароль не переопределен, должен остаться из YAML
	if config.Database.Password != "dbpass" {
		t.Errorf("Expected Database.Password to be 'dbpass', got '%s'", config.Database.Password)
	}
	// Порт сервера должен быть переопределен
	if config.Server.Port != 9090 {
		t.Errorf("Expected Server.Port to be 9090, got %d", config.Server.Port)
	}
	expectedStatuses := []string{"a", "b"}
	if !reflect.DeepEqual(config.Statuses, expectedStatuses) {
		t.Errorf("Expected Statuses to be %v, got %v", expectedStatuses, config.Statuses)
	}
	if config.Enable != true {
		t.Errorf("Expected Enable to be true, got %v", config.Enable)
	}
}

// Проверка значений по умолчанию
func TestConfigManager_NestedConfigDefaultValues(t *testing.T) {
	yamlContent := `
appName: "testapp"
database:
  url: "postgres://localhost:5432/db"
  username: "dbuser"
  password: "dbpass"
server:
  port: 
`
	configPath := createTempYAMLConfig(t, yamlContent)
	defer os.Remove(configPath)

	settings := Settings{
		ConfigFilePath: configPath,
	}

	cm, err := NewConfigManager[TestConfig](settings)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	config := cm.Config()

	// Проверяем значения по умолчанию
	if config.Server.Host != "defaulthost" {
		t.Errorf("Expected Server.Host to be 'defaulthost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8081 {
		t.Errorf("Expected Server.Port to be 8080, got %d", config.Server.Port)
	}
	expectedStatuses := []string{"a", "b", "c", "aa", "ab"}
	if !reflect.DeepEqual(config.Statuses, expectedStatuses) {
		t.Errorf("Expected Statuses to be %v, got %v", expectedStatuses, config.Statuses)
	}
	if config.Enable != true {
		t.Errorf("Expected Enable to be true, got %v", config.Enable)
	}
}
