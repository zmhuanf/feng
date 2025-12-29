package feng

type messageType int

const (
	messageTypeRequest messageType = iota
	messageTypePush
	messageTypeRequestBack
	messageTypePushBack
)

type message struct {
	Route   string      `json:"route"`
	ID      string      `json:"id"`
	Type    messageType `json:"type"`
	Data    string      `json:"data"`
	Success bool        `json:"success"`
}