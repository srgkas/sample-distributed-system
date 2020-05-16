package coordinators

import "time"

type EventAggregator struct {
	listeners map[string][]Listener
}

type EventData struct {
	Name string
	Value float64
	Timestamp time.Time
}

type Listener interface {
	Handle(data EventData)
}

func NewEventAggregator() EventAggregator {
	return EventAggregator{
		listeners: make(map[string][]Listener),
	}
}

func (ea *EventAggregator) AddListener(event string, l Listener) {
	ea.listeners[event] = append(ea.listeners[event], l)
}

func (ea *EventAggregator) FireEvent(event string, data EventData) {
	if ea.listeners[event] == nil {
		return
	}

	for _, handler := range ea.listeners[event] {
		handler.Handle(data)
	}
}