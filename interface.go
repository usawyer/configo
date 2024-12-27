package configo

import (
	"context"

	"github.com/vsysa/configo/notifier"
)

// IConfigManager определяет интерфейс для менеджера конфигурации,
// который загружает и предоставляет доступ к конфигурации типа T.
type IConfigManager[T any] interface {
	// Config возвращает текущую конфигурацию.
	Config() T

	// ChangeCh возвращает канал, по которому можно получать сообщения об изменении конфигурации.
	ChangeCh(ctx context.Context) <-chan notifier.ConfigUpdateMsg[T]
}
