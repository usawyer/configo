# Configo
**Configo**  is a Go library designed for convenient application configuration management. It allows you to read configuration parameters from YAML files and environment variables, with environment variables taking precedence. The library supports configuration validation, automatic reloading upon changes to the configuration file, and provides tools for generating configuration templates.

## Features

- **Load configuration from YAML files and environment variables** : Environment variables override values from the YAML file.

- **Configuration validation** : Ability to define custom validation logic for your configuration.

- **Hot configuration reload** : Automatically monitors changes to the configuration file and updates the configuration in the application.

- **Generate configuration templates** : Automatically create a template YAML file with descriptions and default values.

- **Environment variables help** : Display information about available environment variables and their descriptions.

## Installation


```bash
go get -u github.com/vsysa/configo
```

## Usage

### 1. Define your configuration structure
Create a struct that describes your configuration and implement the `Configurable` interface:

```go
type Configurable interface {
    Validate() error
}
```
Example configuration struct with different data types:**

```go
type DatabaseConfig struct {
    URL      string   `mapstructure:"url" default:"postgres://localhost:5432/db" desc:"Database connection URL"`
    Username string   `mapstructure:"username" desc:"Database username"`
    Password string   `mapstructure:"password" desc:"Database password"`
    Options  []string `mapstructure:"options" default:"sslmode=disable,TimeZone=UTC" desc:"Database connection options"`
}

type ServerConfig struct {
    Host         string `mapstructure:"host" default:"0.0.0.0" desc:"Server host"`
    Port         int    `mapstructure:"port" default:"8080" desc:"Server port"`
    EnableHTTPS  bool   `mapstructure:"enable_https" default:"false" desc:"Enable HTTPS"`
    AllowedIPs   []string `mapstructure:"allowed_ips" default:"127.0.0.1,192.168.1.1" desc:"List of allowed IPs"`
}

type AppConfig struct {
    Database DatabaseConfig `mapstructure:"database" desc:"Database settings"`
    Server   ServerConfig   `mapstructure:"server" desc:"Server settings"`
}

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

In this example:

- `Options` in `DatabaseConfig` is a slice of strings (`[]string`).

- `EnableHTTPS` in `ServerConfig` is a boolean (`bool`).

- `AllowedIPs` in `ServerConfig` is a slice of strings (`[]string`).
2. Create an instance of `ConfigManager`

```go
settings := configo.Settings{
    ConfigFilePath:    "./config.yml", // Name to your YAML configuration file
}

cm, err := configo.NewConfigManager[AppConfig](settings)
```
Configuration is loaded immediately when you create the `ConfigManager` instance.
### 3. Set an error handler (optional)

You can set a custom error handler to handle errors that occur during configuration loading or reloading:


```go
cm.SetErrorHandler(func(err error) {
    log.Printf("Configuration error: %v", err)
})
```

### 4. Use the configuration in your application


```go
config := cm.Config()

fmt.Println("Server running on", config.Server.Host, ":", config.Server.Port)
fmt.Println("Connecting to database at", config.Database.URL)

if config.Server.EnableHTTPS {
    fmt.Println("HTTPS is enabled")
}

fmt.Println("Database options:", config.Database.Options)
fmt.Println("Allowed IPs:", config.Server.AllowedIPs)
```

### 5. Monitor configuration changes

You can subscribe to configuration updates and react to changes:


```go
ctx := context.Background()
go func() {
    for {
        update, ok:= <-cm.ChangeCh(ctx)
        if !ok {
            return
        }
        log.Println("Configuration updated")
        log.Printf("Old value: %+v", update.OldConfig)
        log.Printf("New value: %+v", update.NewConfig)
    }
}()
```

### 6. Generate configuration template and environment variables help


```go
// Generate configuration file template with descriptions
cm.PrintConfigTemplate(true)

// Display information about available environment variables
cm.PrintEnvHelp()
```

## Tags and Annotations

You can use the following tags in your configuration struct:

- `mapstructure`: Specifies the mapping between the struct field and the key in the YAML file or environment variable.

- `default`: Default value for the field if it is not set in the YAML file or environment variable.

- `desc`: Field description, which will be used when generating the configuration template and environment variables help.

- `env`: The name of the environment variable that will override the value from the YAML file.

Example:


```go
type AppConfig struct {
    DebugMode bool `mapstructure:"debug_mode" default:"false" desc:"Enable debug mode" env:"APP_DEBUG_MODE"`
}
```

## Example YAML Configuration


```yaml
# Database settings
database:
  # Database connection URL
  url: "postgres://localhost:5432/db"
  # Database username
  username: "dbuser"
  # Database password
  password: "dbpass"
  # Database connection options
  options:
    - "sslmode=disable"
    - "TimeZone=UTC"

# Server settings
server:
  # Server host
  host: "0.0.0.0"
  # Server port
  port: 8080
  # Enable HTTPS
  enable_https: false
  # List of allowed IPs
  allowed_ips:
    - "127.0.0.1"
    - "192.168.1.1"
```

## Overriding via Environment Variables

Environment variables have a higher priority and override values from the YAML file. Environment variable names are automatically generated based on the configuration structure by replacing dots with underscores and converting to uppercase.

Example:

- Field `database.url` corresponds to the environment variable `DATABASE_URL`.

- Field `server.port` corresponds to the environment variable `SERVER_PORT`.

- Field `server.enable_https` corresponds to the environment variable `SERVER_ENABLE_HTTPS`.
  You can also explicitly specify the environment variable name using the `env` tag.
## Error Handling

Instead of using an error channel, you can set an error handler function to handle errors that occur during configuration loading or reloading:


```go
cm.SetErrorHandler(func(err error) {
    log.Printf("Configuration error: %v", err)
})
```

## Generating Configuration Template

You can automatically generate a configuration file template based on your configuration structure:


```go
cm.PrintConfigTemplate(true)
```

Example output:


```yaml
# Database settings
database:
    # Database connection URL
    url: "postgres://localhost:5432/db"
    # Database username
    username: ""
    # Database password
    password: ""
    # Database connection options
    options:
        - "sslmode=disable"
        - "TimeZone=UTC"

# Server settings
server:
    # Server host
    host: "0.0.0.0"
    # Server port
    port: 8080
    # Enable HTTPS
    enable_https: false
    # List of allowed IPs
    allowed_ips:
        - "127.0.0.1"
        - "192.168.1.1"
```

## Environment Variables Help

To display information about available environment variables:


```go
cm.PrintEnvHelp()
```

Example output:


```
Environment Variables:
	DATABASE_URL         string     Database connection URL (default: postgres://localhost:5432/db)
	DATABASE_USERNAME    string     Database username (default: )
	DATABASE_PASSWORD    string     Database password (default: )
	DATABASE_OPTIONS     []string   Database connection options (default: [sslmode=disable TimeZone=UTC])
	SERVER_HOST          string     Server host (default: 0.0.0.0)
	SERVER_PORT          int        Server port (default: 8080)
	SERVER_ENABLE_HTTPS  bool       Enable HTTPS (default: false)
	SERVER_ALLOWED_IPS   []string   List of allowed IPs (default: [127.0.0.1 192.168.1.1])
```

## Hot Configuration Reload Support

The library automatically monitors changes to the configuration file and updates the configuration in your application. You can subscribe to the change channel and react to updates.

## Configuration Validation
Implement the `Validate()` method in your configuration struct to check the correctness of settings. The method should return an error if the configuration is invalid.

```go
func (c *AppConfig) Validate() error {
    if c.Server.Port <= 0 || c.Server.Port > 65535 {
        return fmt.Errorf("server port must be between 1 and 65535")
    }
    if c.Server.EnableHTTPS && c.Server.Port != 443 {
        return fmt.Errorf("when HTTPS is enabled, server port must be 443")
    }
    if len(c.Server.AllowedIPs) == 0 {
        return fmt.Errorf("allowed IPs cannot be empty")
    }
    // Other checks...
    return nil
}
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Authors

- **Vladislav Sysalov** - [vsysa](https://github.com/vsysa)

## Contributing

We welcome your suggestions and improvements! Please open issues and pull requests.