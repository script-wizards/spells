package engine

import "sync"

type Event interface {
	Type() string
}

type TurnAdvanced struct {
	SessionID int64
	OldTurn   int64
	NewTurn   int64
	Delta     int64
}

func (e TurnAdvanced) Type() string {
	return "TurnAdvanced"
}

type EventHandler func(Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (bus *EventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}

func (bus *EventBus) Emit(event Event) {
	bus.mu.RLock()
	handlers := bus.handlers[event.Type()]
	bus.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}
