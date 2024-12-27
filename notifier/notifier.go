package notifier

import (
	"context"
	"sync"
)

// ConfigUpdateMsg представляет сообщение об обновлении конфигурации,
// содержащее старую и новую версии конфигурации.
type ConfigUpdateMsg[T any] struct {
	OldConfig T
	NewConfig T
}

type ConfigUpdateNotifier[T any] struct {
	mu          sync.RWMutex
	subscribers map[chan ConfigUpdateMsg[T]]struct{}
}

// NewEventBus создает новый eventBus.
func NewConfigUpdateNotifier[T any]() *ConfigUpdateNotifier[T] {
	return &ConfigUpdateNotifier[T]{
		subscribers: make(map[chan ConfigUpdateMsg[T]]struct{}),
	}
}

// Subscribe позволяет подписчику получать события. Возвращает канал, через который будут получены события.
func (r *ConfigUpdateNotifier[T]) Subscribe(ctx context.Context) <-chan ConfigUpdateMsg[T] {
	ch := make(chan ConfigUpdateMsg[T], 1) // Используем буферизированный канал для предотвращения блокировки
	r.mu.Lock()
	r.subscribers[ch] = struct{}{}
	r.mu.Unlock()

	go func() {
		<-ctx.Done()
		r.mu.Lock()
		delete(r.subscribers, ch)
		close(ch)
		r.mu.Unlock()
	}()

	return ch
}

// Publish публикует событие всем подписчикам.
func (r *ConfigUpdateNotifier[T]) NewEvent(msg ConfigUpdateMsg[T]) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for ch := range r.subscribers {
		select {
		case ch <- msg: // Отправляем событие, если канал готов принять сообщение
		default: // Пропускаем, если в канале уже есть сообщение
		}
	}
}
