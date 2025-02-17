# Configo
**Configo**  is a Go library that provides a convenient wrapper around [Viper](https://github.com/spf13/viper)  for managing application configuration. It can read parameters from a YAML file and environment variables, with environment variables taking precedence. The library supports configuration validation, automatic “hot” reloading of the configuration file, and offers utilities for generating a YAML configuration template and environment variable help.
## Features

- **Load configuration from YAML files and environment variables**
  Environment variables override YAML values.

- **Configuration validation**
  By implementing a `Validate()` method in your struct, you can check the correctness of the loaded configuration.

- **Hot reload**
  Uses Viper’s fsnotify-based watch to automatically reload configurations if the YAML file changes.

- **YAML template generation**
  Generates a commented YAML file with default values, based on your struct definitions.

- **Environment variable help**
  Shows available environment variables, their types, defaults, and the corresponding struct field key (BindKey).

## Installation


```bash
go get -u github.com/vsysa/configo
```

## Quick Start

### 1. Define Your Configuration Struct
Create a struct describing your configuration. Optionally, implement a `Validate()` method to verify correctness.

```go
type DatabaseConfig struct {
    URL      string   `mapstructure:"url" default:"postgres://localhost:5432/db" help:"Database connection URL"`
    Username string   `mapstructure:"username" help:"Database username"`
    Password string   `mapstructure:"password" help:"Database password"`
    Options  []string `mapstructure:"options" default:"sslmode=disable,TimeZone=UTC" help:"Database connection options"`
}

type ServerConfig struct {
    Host        string   `mapstructure:"host" default:"0.0.0.0" help:"Server host"`
    Port        int      `mapstructure:"port" default:"8080"     help:"Server port"`
    EnableHTTPS bool     `mapstructure:"enable_https" default:"false" help:"Enable HTTPS?"`
    AllowedIPs  []string `mapstructure:"allowed_ips" default:"127.0.0.1,192.168.1.1" help:"List of allowed IPs"`
}

// Example of using env:"srv" on the Server field, so environment variables
// for its fields will have the SRV_ prefix.
type AppConfig struct {
    Database DatabaseConfig `mapstructure:"database" help:"Database settings"`
    Server   ServerConfig   `mapstructure:"server"   help:"Server settings" env:"srv"`
}

// Optional validation method
func (c *AppConfig) Validate() error {
    if c.Server.Port <= 0 {
        return fmt.Errorf("server port must be a positive number")
    }
    if c.Database.URL == "" {
        return fmt.Errorf("database URL cannot be empty")
    }
    return nil
}
```
2. Create a `ConfigManager` Instance (Wrapper Around Viper)Use “options” (`Option` functions) for initialization:

```go
cm, err := configo.NewConfigManager[AppConfig](
    configo.WithConfigFilePath[AppConfig]("./config.yml"),
    configo.WithErrorHandler[AppConfig](func(err error) {
        log.Printf("Configuration error: %v", err)
    }),
)
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### 3. Working With the Configuration


```go
func main() {
    // Retrieve the current configuration
    config := cm.Config()

    fmt.Println("Server listening on", config.Server.Host, ":", config.Server.Port)
    fmt.Println("Database URL:", config.Database.URL)
    
    // Check if HTTPS is enabled
    if config.Server.EnableHTTPS {
        fmt.Println("HTTPS is enabled!")
    }

    // Subscribe to changes (hot reload)
    ctx := context.Background()
    go func() {
        for update := range cm.ChangeCh(ctx) {
            log.Println("Configuration updated!")
            log.Printf("Old config: %+v", update.OldConfig)
            log.Printf("New config: %+v", update.NewConfig)
        }
    }()

    // Block or continue doing other stuff...
    select {}
}
```

## Tags Overview
Configo relies on specific tags within struct fields to determine how to parse and interpret configuration values. Under the hood, it leverages [Viper](https://github.com/spf13/viper) , but provides additional conveniences for default values, environment variable mappings, and documentation.
Below are all the supported tags, each with detailed rules and examples.


---

1. `mapstructure:"..."`
- **Purpose** : Defines the key by which Viper will map data from YAML (and from environment variables if `env:"..."` is not specified).

- **Default Behavior** :
  - If `mapstructure` is **not**  present, the field name (in lowercase) is used as the key.
    E.g., a Go field named `MyField` becomes `myfield`.

  - If `mapstructure:"-"`, the field is ignored for YAML/environment variable binding.

- **Example** :

```go
type ServerConfig struct {
    Port int `mapstructure:"port"`
    // ...
}
```

If the YAML looks like:


```yaml
server:
  port: 8080
```
Then `ServerConfig.Port` will be populated with `8080`.


---

2. `env:"..."`
- **Purpose** : Explicitly sets or disables the name of the environment variable(s).

- **Rules** :
  1. **If**  `env:"-"`, **then**  no environment variables are generated for this field **or**  (if placed on a struct) for any of its child fields.
  - *Important*: Placing `env:"-"` on a struct effectively hides **all**  of its fields from environment variable overrides.

  2. If `env:"MY_VAR"`, that becomes the environment variable name (after converting to uppercase), e.g. `MY_VAR`.

  3. If `env` is **not**  set, but `mapstructure:"foo"` is set, the environment variable name becomes `FOO` (uppercase).

  4. If neither `env` nor `mapstructure` is present, the environment variable name is derived from the field name (uppercase).

- **Examples** :

```go
// 1) Simple override with env
type ServerConfig struct {
    Port int `mapstructure:"port" env:"srv_port"`
}
// => Environment variable: SRV_PORT

// 2) Completely disable environment variables for a field
type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Key  string `mapstructure:"key" env:"-"` // env is disabled here
}
// => Only SERVER_PORT is recognized from environment, SERVER_KEY is NOT recognized

// 3) Disable environment variables for the entire struct
type SecretConfig struct {
    Password string `mapstructure:"password" default:"secret"`
}

type AppConfig struct {
    Secret SecretConfig `mapstructure:"secret" env:"-"`
}
// => None of the fields under SecretConfig are exposed to environment variables
```


---

3. `default:"..."`
- **Purpose** : Specifies a default value if the field is not set via YAML or environment variables.

- **Supported Types** :
  - Primitives (e.g., `string`, `bool`, numeric types)

  - **Slices**  of primitives (via a JSON-like array or comma-separated list)

  - **Maps**  of primitive keys/values in JSON form (e.g. `{"key":"value"}`)

- **Rules for Slices** :
  1. If the default value is valid JSON (e.g., `"[\"val1\", \"val2\"]"`), it will be parsed as JSON.

  2. Otherwise, the default can be written as a comma-separated string (e.g., `"val1,val2"`), which is split into `[]string{"val1", "val2"}`.

  3. If a slice’s element type is not primitive (e.g., slice of structs), automatic parsing of defaults will **not**  work (the default is ignored).

- **Examples** :

```go
type ServerConfig struct {
    Host       string   `mapstructure:"host" default:"0.0.0.0"`
    Port       int      `mapstructure:"port" default:"8080"`
    AllowedIPs []string `mapstructure:"allowed_ips" default:"127.0.0.1,192.168.1.1"`
    // or using a JSON array:
    // default:"[\"127.0.0.1\",\"192.168.1.1\"]"
}
```


```go
type ExampleConfig struct {
    // A map with a JSON default
    Settings map[string]string `mapstructure:"settings" default:"{\"env\":\"prod\",\"region\":\"us-east\"}"`
}
```


---

4. `help:"..."`
- **Purpose** : Provides a description/comment for the field (or struct) used when:
  1. Generating YAML templates (comments next to the field)

  2. Printing environment variable help (inline or tabular)

- **Example** :

```go
type ServerConfig struct {
    Port int `mapstructure:"port" default:"8080" help:"The port on which the server listens"`
}
```
**YAML template**  example:

```yaml
server:
  port: 8080  # The port on which the server listens
```
**Environment variable help**  example (inline format):

```php
SERVER_PORT [default=8080] [BindKey=server.port] [ValueType=int] # The port on which the server listens
```


---


### Tag Precedence and Interaction

1. **`env:"-"`**  has the highest priority in terms of disabling environment variables:
- If set on a struct, **all**  nested fields are skipped.

- If set on an individual field, only that field is skipped.

2. If `env:"SOMETHING"` is explicitly set, that name (uppercased) is used for the environment variable.

3. Otherwise, if `mapstructure:"foo"` is present, the environment variable becomes `FOO` (uppercase).

4. If no `env` or `mapstructure` is present, the Go field name (uppercase) is used.

5. The `default:"..."` tag is used if no value is found in YAML or environment variables.

6. The `help:"..."` tag is purely for documentation (YAML template generation and environment variable help output).


[//]: # ( need to check)
[//]: # (> **Important Note** : If you use `mapstructure:"-"` on a field, it is ignored by Viper entirely &#40;neither YAML nor environment variables can set it&#41;. This is distinct from using `env:"-"`, which only disables environment variables but does not affect YAML binding &#40;as long as `mapstructure` is something other than `-`&#41;.)

---


## Generating a YAML Template


```go
import "github.com/vsysa/configo/internal/helper"

// Generate a YAML template (with help comments and default values)
fmt.Println(helper.GenerateYAMLTemplate(AppConfig{}))
```

For example (simplified):


```yaml
database:
  url: "postgres://localhost:5432/db"  # Database connection URL
  username: null                       # Database username
  password: null                       # Database password
  options:
    - sslmode=disable
    - TimeZone=UTC                     # Database connection options

srv:
  host: "0.0.0.0"                      # Server host
  port: 8080                           # Server port
  enable_https: false                  # Enable HTTPS?
  allowed_ips:
    - 127.0.0.1
    - 192.168.1.1                      # List of allowed IPs
```

## Environment Variable Help


```go
import "github.com/vsysa/configo/internal/helper"

// Print a list of environment variables in different formats
fmt.Println(helper.GenerateEnvHelp(AppConfig{}, helper.Inline))
fmt.Println(helper.GenerateEnvHelp(AppConfig{}, helper.AsciiTable))
fmt.Println(helper.GenerateEnvHelp(AppConfig{}, helper.MarkdownTable))
```
Example in `Inline` format:

```csharp
SRV_HOST [default=0.0.0.0] # Server host
SRV_PORT [default=8080] # Server port
SRV_ENABLE_HTTPS [default=false] # Enable HTTPS?
SRV_ALLOWED_IPS [default=sslmode=disable,TimeZone=UTC] # List of allowed IPs
DATABASE_URL [default=postgres://localhost:5432/db] # Database connection URL
...
```

## Example YAML Configuration


```yaml
# Database settings
database:
  url: "postgres://localhost:5432/db"
  username: "dbuser"
  password: "dbpass"
  options:
    - "sslmode=disable"
    - "TimeZone=UTC"

# Server settings
server:
  host: "0.0.0.0"
  port: 8080
  enable_https: false
  allowed_ips:
    - "127.0.0.1"
    - "192.168.1.1"
```

## Overriding Via Environment Variables

Environment variables take precedence over YAML. The variable names are generated as follows:

1. If `env:"..."` is set, that value (in uppercase) is used.

2. If `env` is missing but `mapstructure:"..."` is set, it’s converted to uppercase.

3. Otherwise, the variable name is derived from the field name in uppercase.
   When structs are nested, prefixes are concatenated with `_`. For example, if `ServerConfig` has `env:"srv"`, and the `Host` field does not override `env`, the resulting variable is `SRV_HOST`.
## Validation
If your struct implements `Validate() error`, that method is called after loading from YAML/environment variables and before making the configuration available to the application. If validation fails, an error is returned or the provided `errorHandler` is triggered.

```go
func (c *AppConfig) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
}
```

## Error Handling

Instead of an error channel, you can set your own error handler:


```go
cm.SetErrorHandler(func(err error) {
    log.Printf("Configuration error: %v", err)
})
```

## License
This project is licensed under the MIT License. See the [LICENSE](https://chatgpt.com/c/LICENSE)  file for details.
## Authors

- **Vladislav Sysalov**  — [vsysa](https://github.com/vsysa)
