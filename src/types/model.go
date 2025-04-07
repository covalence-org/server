package types

import (
	"errors"
	"fmt"
	"net/url"
)

// ========================= Name =========================

type Name struct {
	raw string
}

func (s Name) Complete() bool {
	return s.raw != ""
}

func (s Name) String() string {
	return s.raw
}

// isValidModelName checks if a model name contains only valid characters
func isValidName(name string) bool {
	if len(name) < 1 || len(name) > 64 {
		return false
	}

	// Allow alphanumeric, dash, underscore, and dot
	for _, r := range name {
		if !(('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}

	return true
}

func NewName(value string) (Name, error) {
	if value == "" {
		return Name{}, errors.New("empty name")
	}
	if !isValidName(value) {
		return Name{}, errors.New("invalid name")
	}
	return Name{value}, nil
}

// ========================= APIURL =========================

type APIURL struct {
	raw string
}

func (s APIURL) Complete() bool {
	return s.raw != ""
}

func (s APIURL) String() string {
	return s.raw
}

func isValidAPIURL(value string) bool {
	parsedURL, err := url.ParseRequestURI(value)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func NewAPIURL(value string) (APIURL, error) {
	if value == "" {
		return APIURL{}, errors.New("APIURL cannot be empty")
	}
	if !isValidAPIURL(value) {
		return APIURL{}, fmt.Errorf("APIURL '%s' is invalid", value)
	}
	return APIURL{value}, nil
}

// ========================= ModelProvider =========================

type ModelProvider struct {
	raw string
}

func (s ModelProvider) Complete() bool {
	return s.raw != ""
}

func (s ModelProvider) String() string {
	return s.raw
}

func isValidModelProvider(value string) bool {
	validTypes := map[string]struct{}{
		"openai":    {},
		"anthropic": {},
		"google":    {},
		"meta":      {},
		"custom":    {},
	}
	_, exists := validTypes[value]
	return exists
}

func NewModelProvider(value string) (ModelProvider, error) {
	if value == "" {
		return ModelProvider{}, errors.New("ModelProvider cannot be empty")
	}
	if !isValidModelProvider(value) {
		return ModelProvider{}, fmt.Errorf("ModelProvider '%s' is invalid", value)
	}
	return ModelProvider{value}, nil
}
