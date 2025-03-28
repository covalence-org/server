package types

import (
	"errors"
)

type InternalModel struct {
	raw string
}

func (s InternalModel) Complete() bool {
	return s.raw != ""
}

func (s InternalModel) String() string {
	return s.raw
}

func NewInternalModel(value string) (InternalModel, error) {
	if value == "" {
		return InternalModel{}, errors.New("Model cannot be empty")
	}
	return InternalModel{value}, nil
}

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
		return InternalModelType{}, errors.New("Model cannot be empty")
	}
	if !isValidInternalModelType(value) {
		return InternalModelType{}, errors.New("Invalid model type")
	}
	return InternalModelType{value}, nil
}
