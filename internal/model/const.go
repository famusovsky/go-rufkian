package model

import "errors"

var (
	ErrNoHistoryFound  = errors.New("NO_HISTORY_FOUND")
	ErrEmptyDialog     = errors.New("EMPTY_DIALOG")
	ErrWrongBodyFormat = errors.New("WRONG_BODY_FORMAT")
	ErrEmptyUserID     = errors.New("EMPTY_USER_ID")
	ErrEmptyKey        = errors.New("EMPTY_KEY")
	ErrEmptyInput      = errors.New("EMPTY_INPUT")
)

const (
	AssistantRole Role = "assistant"
	SystemRole    Role = "system"
	ToolRole      Role = "tool"
	UserRole      Role = "user"
)

const (
	MistralSmall      MistralModel = "mistral-small-latest"
	MistralLarge      MistralModel = "mistral-large-latest"
	MistralModeration MistralModel = "mistral-moderation-latest"
)
