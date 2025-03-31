package types

import (
	"errors"
)

// ======== Internal Model Types ==========

type InternalModelType struct {
	raw string
}

func (s InternalModelType) Complete() bool {
	return s.raw != ""
}

func (s InternalModelType) String() string {
	return s.raw
}

func isValidInternalModelType(value string) bool {
	return value == "text-classification" || value == "image-classification"
}

func NewInternalModelType(value string) (InternalModelType, error) {
	if value == "" {
		return InternalModelType{}, errors.New("model cannot be empty")
	}
	if !isValidInternalModelType(value) {
		return InternalModelType{}, errors.New("invalid model type")
	}
	return InternalModelType{value}, nil
}
