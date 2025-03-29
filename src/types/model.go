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

// ========================= ApiUrl =========================

type ApiUrl struct {
	raw string
}

func (s ApiUrl) Complete() bool {
	return s.raw != ""
}

func (s ApiUrl) String() string {
	return s.raw
}

func isValidApiUrl(value string) bool {
	parsedUrl, err := url.ParseRequestURI(value)
	return err == nil && parsedUrl.Scheme != "" && parsedUrl.Host != ""
}

func NewApiUrl(value string) (ApiUrl, error) {
	if value == "" {
		return ApiUrl{}, errors.New("ApiUrl cannot be empty")
	}
	if !isValidApiUrl(value) {
		return ApiUrl{}, fmt.Errorf("ApiUrl '%s' is invalid", value)
	}
	return ApiUrl{value}, nil
}
