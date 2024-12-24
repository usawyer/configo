package configo

type Option[T Configurable] func(*ConfigManager[T])

func WithConfigFilePath[T Configurable](path string) Option[T] {
	return func(cm *ConfigManager[T]) {
		cm.configFilePath = path
	}
}

func WithErrorHandler[T Configurable](handler func(error)) Option[T] {
	return func(cm *ConfigManager[T]) {
		cm.errorHandler = handler
	}
}
