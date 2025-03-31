package types

import (
	"errors"
)

type ModelID struct {
	raw string
}

func (s ModelID) Complete() bool {
	return s.raw != ""
}

func (s ModelID) String() string {
	return s.raw
}

func NewModelID(value string) (ModelID, error) {
	if value == "" {
		return ModelID{}, errors.New("Model cannot be empty")
	}
	return ModelID{value}, nil
}
