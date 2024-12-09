package configo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/vsysa/configo/internal/helper"
	"github.com/vsysa/configo/internal/parser"
)

const (
	DefaultConfigPath = "./config.yml"
)

type Settings struct {
	ConfigFilePath string
}

type ConfigManager[T Configurable] struct {
	config *T

	settings Settings

	configUpdateNotifier *ConfigUpdateNotifier[T]
	updateMu             sync.RWMutex
	errorHandler         func(error)
	v                    *viper.Viper
	configTree           *parser.ConfigNode
}

func MustNewConfigManager[T Configurable](managerSettings Settings) *ConfigManager[T] {
	out, err := NewConfigManager[T](managerSettings)
	if err != nil {
		panic(err)
	}
	return out
}

func NewConfigManager[T Configurable](managerSettings Settings) (*ConfigManager[T], error) {
	r := &ConfigManager[T]{
		settings:             managerSettings,
		configUpdateNotifier: NewConfigUpdateNotifier[T](),
		errorHandler:         helper.DefaultHandleError,
		v:                    viper.New(),
	}

	var cfg T
	var err error
	r.configTree, err = parser.ParseConfigStruct(cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing config struct: %w", err)
	}

	configPath := managerSettings.ConfigFilePath
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	r.setupViper(configPath)
	r.setupWatcher()

	if _, err := r.updateConfig(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *ConfigManager[T]) Config() T {
	if r.config == nil {
		panic("ConfigManager has not been initialized")
	}
	r.updateMu.RLock()
	defer r.updateMu.RUnlock()

	return *r.config
}

func (r *ConfigManager[T]) ChangeCh(ctx context.Context) <-chan ConfigUpdateMsg[T] {
	return r.configUpdateNotifier.Subscribe(ctx)
}

func (r *ConfigManager[T]) SetErrorHandler(handler func(error)) {
	r.updateMu.Lock()
	defer r.updateMu.Unlock()
	r.errorHandler = handler
}

func (r *ConfigManager[T]) updateConfig() (*T, error) {
	newConfig, err := r.loadConfig()
	if err != nil {
		return nil, err
	}
	r.updateMu.Lock()
	r.config = newConfig
	r.updateMu.Unlock()
	return newConfig, nil
}

func (r *ConfigManager[T]) loadConfig() (*T, error) {
	Viper := r.v

	if err := Viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg T
	if err := Viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("Unable to decode into struct: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("Validation error: %v", err)
	}

	return &cfg, nil
}

func (r *ConfigManager[T]) setupViper(configPath string) {
	Viper := r.v

	Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	Viper.AutomaticEnv()

	Viper.SetConfigFile(configPath)

	for _, item := range r.configTree.GetAllLeaves() {
		if item.ConfigDescription.Default.IsExist {
			Viper.SetDefault(strings.Join(item.GetFullPathParts(), "."), item.ConfigDescription.Default.Value)
		}
		// по дефолту viper сам связывает названия с переменными env, но мы это делаем для прозрачности
		// и на случай если пользователь изменил название переменной. если пользователь написал env:"-",
		// то автоматическое связывание все равно произойдет
		// связываем переменную с названием env
		if envName, exist := item.GetEnv(); exist {
			Viper.BindEnv(strings.Join(item.GetFullPathParts(), "."), envName)
		}
	}
}

func (r *ConfigManager[T]) setupWatcher() {
	Viper := r.v
	Viper.OnConfigChange(func(e fsnotify.Event) {
		//fmt.Println("Config file changed:", e.Name)
		oldConfig := r.Config()
		newConfig, err := r.updateConfig()
		if err != nil {
			r.errorHandler(fmt.Errorf("Unable to load config on update: %v", err))
			return
		}

		r.configUpdateNotifier.NewEvent(ConfigUpdateMsg[T]{
			OldConfig: oldConfig,
			NewConfig: *newConfig,
		})
	})

	Viper.WatchConfig()
}

var _ IConfigManager[Configurable] = &ConfigManager[Configurable]{}
