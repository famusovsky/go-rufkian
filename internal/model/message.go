package model

type Role string

type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Prefix  bool   `json:"prefix,omitempty"`
	// TODO tool_calls
}

func (msg *Message) Empty() bool {
	if msg == nil {
		return true
	}
	return msg.Role == "" && msg.Content == "" && !msg.Prefix
}

type Messages []Message

type MistralModel string
