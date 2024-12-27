package notifier

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Структуры для тестирования
type MockConfig struct {
	Value string
}

func (MockConfig) Validate() error {
	return nil
}

// Тест на успешную подписку и получение событий
func TestConfigUpdateNotifier_SubscribeAndPublish(t *testing.T) {
	notifier := NewConfigUpdateNotifier[MockConfig]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подписываемся на события
	subscriber := notifier.Subscribe(ctx)

	oldValue := "OldConfig"
	newValue := "NewConfig"

	// Публикуем событие
	testConfig := ConfigUpdateMsg[MockConfig]{OldConfig: MockConfig{Value: oldValue}, NewConfig: MockConfig{Value: newValue}}
	notifier.NewEvent(testConfig)

	// Проверяем, что подписчик получил событие
	select {
	case msg := <-subscriber:
		assert.Equal(t, oldValue, msg.OldConfig.Value)
		assert.Equal(t, newValue, msg.NewConfig.Value)
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

// Тест на удаление подписчиков после завершения контекста
func TestConfigUpdateNotifier_SubscriberCleanup(t *testing.T) {
	notifier := NewConfigUpdateNotifier[MockConfig]()
	ctx, cancel := context.WithCancel(context.Background())

	// Подписываемся на события
	subscriber := notifier.Subscribe(ctx)

	// Отменяем контекст, чтобы удалить подписчика
	cancel()

	// даем возможность горутине закрыть канал
	runtime.Gosched()

	// Публикуем событие
	testConfig := ConfigUpdateMsg[MockConfig]{NewConfig: MockConfig{Value: "TestValue"}}
	notifier.NewEvent(testConfig)

	// Проверяем, что подписчик не получил событие, так как контекст был отменен
	_, ok := <-subscriber
	assert.False(t, ok, "Expected channel to be closed, but it's still open")
}

// Тест на корректную работу при множественных подписчиках
func TestConfigUpdateNotifier_MultipleSubscribers(t *testing.T) {
	notifier := NewConfigUpdateNotifier[MockConfig]()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем двух подписчиков
	subscriber1 := notifier.Subscribe(ctx)
	subscriber2 := notifier.Subscribe(ctx)

	// Публикуем событие
	testConfig := ConfigUpdateMsg[MockConfig]{NewConfig: MockConfig{Value: "TestValue"}}
	notifier.NewEvent(testConfig)

	// Проверяем, что оба подписчика получили событие
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		select {
		case msg := <-subscriber1:
			assert.Equal(t, testConfig.NewConfig.Value, msg.NewConfig.Value)
		case <-time.After(1 * time.Second):
			t.Error("Subscriber 1 timeout waiting for event")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case msg := <-subscriber2:
			assert.Equal(t, testConfig.NewConfig.Value, msg.NewConfig.Value)
		case <-time.After(1 * time.Second):
			t.Error("Subscriber 2 timeout waiting for event")
		}
	}()

	wg.Wait()
}
