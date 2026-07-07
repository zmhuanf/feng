package protocol

type MessageType int

const (
	MessageTypeRequest MessageType = iota
	MessageTypePush
	MessageTypeRequestBack
	MessageTypePushBack
)

type Message struct {
	Route   string      `json:"route"`
	ID      string      `json:"id"`
	Type    MessageType `json:"type"`
	Data    string      `json:"data"`
	Success bool        `json:"success"`
}
