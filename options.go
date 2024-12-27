package configo

type Option[T any] func(*ConfigManager[T])

func WithConfigFilePath[T any](path string) Option[T] {
	return func(cm *ConfigManager[T]) {
		cm.configFilePath = path
	}
}

func WithErrorHandler[T any](handler func(error)) Option[T] {
	return func(cm *ConfigManager[T]) {
		cm.errorHandler = handler
	}
}
