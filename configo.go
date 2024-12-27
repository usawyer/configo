package configo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/vsysa/configo/internal/parser/defaultValues"
	"github.com/vsysa/configo/internal/parser/env"
	"github.com/vsysa/configo/notifier"
)

const (
	DefaultConfigPath = "./config.yml"
)

var (
	ConfigParsingError error = errors.New("error parsing config struct")
)

type ConfigManager[T any] struct {
	config *T

	configFilePath string

	configUpdateNotifier *notifier.ConfigUpdateNotifier[T]
	updateMu             sync.RWMutex
	errorHandler         func(error)
	v                    *viper.Viper
}

func MustNewConfigManager[T any](opts ...Option[T]) *ConfigManager[T] {
	out, err := NewConfigManager[T](opts...)
	if err != nil {
		panic(err)
	}
	return out
}

func NewConfigManager[T any](opts ...Option[T]) (*ConfigManager[T], error) {
	r := &ConfigManager[T]{
		configFilePath:       DefaultConfigPath,
		configUpdateNotifier: notifier.NewConfigUpdateNotifier[T](),
		errorHandler: func(err error) {
			log.Printf("ConfigManager error: %v", err)
		},
		v: viper.New(),
	}

	for _, opt := range opts {
		opt(r)
	}

	err := r.setupViper(r.configFilePath)
	if err != nil {
		return nil, err
	}
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

func (r *ConfigManager[T]) ChangeCh(ctx context.Context) <-chan notifier.ConfigUpdateMsg[T] {
	return r.configUpdateNotifier.Subscribe(ctx)
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

	if err := callValidateIfExists(cfg); err != nil {
		return nil, fmt.Errorf("Validation error: %w", err)
	}

	return &cfg, nil
}

func (r *ConfigManager[T]) setupViper(configPath string) error {
	Viper := r.v

	Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	//Viper.AutomaticEnv()

	Viper.SetConfigFile(configPath)

	var configStruct T
	defaults, err := defaultValues.GetDefaultValues(configStruct)
	if err != nil {
		return fmt.Errorf("%w: %w", ConfigParsingError, err)
	}
	for _, v := range defaults {
		Viper.SetDefault(v.BindKey, v.DefaultValue)
	}

	for _, v := range env.GetEnvs(configStruct) {
		err := Viper.BindEnv(v.BindKey, v.EnvVar)
		if err != nil {
			return fmt.Errorf("error binding env var: %w", err)
		}
	}

	return nil
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

		r.configUpdateNotifier.NewEvent(notifier.ConfigUpdateMsg[T]{
			OldConfig: oldConfig,
			NewConfig: *newConfig,
		})
	})

	Viper.WatchConfig()
}

func callValidateIfExists(in interface{}) error {

	// Ищем метод Validate
	method := reflect.ValueOf(in).MethodByName("Validate")
	if !method.IsValid() {
		return nil
	}

	// Проверяем сигнатуру метода
	methodType := method.Type()
	if methodType.NumIn() != 0 || methodType.NumOut() != 1 {
		return nil
	}

	if methodType.Out(0).Name() != "error" {
		return nil
	}

	results := method.Call(nil)
	if err, ok := results[0].Interface().(error); ok {
		return err
	}

	return nil
}

var _ IConfigManager[any] = &ConfigManager[any]{}
