package types

import (
	"errors"
	"fmt"
)

type Message struct {
	role    string
	content string
}

func (s Message) Complete() bool {
	return s.role != "" && s.content != ""
}

func (s Message) ToMap() map[string]string {
	return map[string]string{
		"role":    s.role,
		"content": s.content,
	}
}

func isValidRole(value string) bool {
	return value == "user" || value == "assistant"
}

func isValidContent(value string) bool {
	return value != ""
}

func NewMessage(role string, content string) (Message, error) {
	if role == "" {
		return Message{}, errors.New("role cannot be empty")
	}
	if content == "" {
		return Message{}, errors.New("content cannot be empty")
	}

	if !isValidRole(role) {
		return Message{}, fmt.Errorf("role '%s' is invalid", role)
	}
	if !isValidContent(content) {
		return Message{}, fmt.Errorf("content '%s' is invalid", content)
	}

	return Message{role, content}, nil
}

func NewMessageFromJson(object interface{}) (Message, error) {
	messageObject, ok := object.(map[string]interface{})
	if !ok {
		return Message{}, fmt.Errorf("invalid message format")
	}

	message, err := NewMessage(messageObject["role"].(string), messageObject["content"].(string))
	if err != nil {
		return Message{}, fmt.Errorf("failed to parse message: %v", err)
	}

	return message, nil

}
