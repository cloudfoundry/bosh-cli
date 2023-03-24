package partybus

type EventType string

type Event struct {
	Type   EventType
	Source interface{}
	Value  interface{}
	Error  error
}
