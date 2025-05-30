package model

import (
	"slices"
	"time"
)

type MistralModel string

type Role string

type Message struct {
	Role        Role    `json:"role"`
	Content     string  `json:"content"`
	Translation *string `json:"translation,omitempty"`
	Prefix      bool    `json:"prefix,omitempty"`
}

func (msg *Message) Empty() bool {
	if msg == nil {
		return true
	}
	return msg.Role == "" && msg.Content == "" && !msg.Prefix
}

type Messages []Message

func (msgs Messages) WithoutSystem() Messages {
	return slices.DeleteFunc(msgs, func(msg Message) bool {
		return msg.Role != UserRole && msg.Role != AssistantRole
	})
}

type Dialog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Messages  Messages  `json:"messages"`
	StartTime time.Time `json:"start_time"`
	UpdatedAt time.Time `json:"-"`
	DurationS int       `json:"duration_s"`
}

type Dialogs []Dialog
