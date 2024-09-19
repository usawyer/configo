package configManager

import "context"

// Configurable представляет собой интерфейс для конфигурационных структур,
// которые могут быть проверены на корректность.
type Configurable interface {
	// Validate проверяет корректность конфигурации и возвращает ошибку, если она некорректна.
	Validate() error
}

// ConfigUpdateMsg представляет сообщение об обновлении конфигурации,
// содержащее старую и новую версии конфигурации.
type ConfigUpdateMsg[T Configurable] struct {
	OldConfig T
	NewConfig T
}

// IConfigManager определяет интерфейс для менеджера конфигурации,
// который загружает и предоставляет доступ к конфигурации типа T.
type IConfigManager[T Configurable] interface {
	// Config возвращает текущую конфигурацию.
	Config() T

	// ChangeCh возвращает канал, по которому можно получать сообщения об изменении конфигурации.
	ChangeCh(ctx context.Context) <-chan ConfigUpdateMsg[T]

	// SetErrorHandler Устанавливает обработчик ошибок
	SetErrorHandler(handler func(error))
}

// IConfigInspector определяет интерфейс для анализа конфигурационной структуры.
type IConfigInspector[T Configurable] interface {
	// PrintConfigTemplate выводит шаблон конфигурационного файла.
	// Если printConfigWithDescription истинно, выводит также описания полей.
	PrintConfigTemplate(printConfigWithDescription bool)

	// PrintEnvHelp выводит справочную информацию о переменных окружения, используемых в конфигурации.
	PrintEnvHelp()
}
