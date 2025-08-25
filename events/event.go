package events

type Event struct {
	ID               string
	Type             EventType
	Timestamp        uint64
	ParserIdentifier int64
	ParserName       string
	Message          string
	Attributes       map[string]string
}

// NewEvent creates a new event
func NewEvent(id string, e_type EventType, ts uint64, parser_identifier int64, parser_name string, message string) *Event {
	return &Event{
		ID:               id,
		Type:             e_type,
		Timestamp:        ts,
		ParserIdentifier: parser_identifier,
		ParserName:       parser_name,
		Message:          message,
		Attributes:       make(map[string]string),
	}
}

// AddAttribute adds a new attribute to the event
func (e *Event) AddAttribute(key string, value string) {
	e.Attributes[key] = value
}
