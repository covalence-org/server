package types

import (
	"errors"
	"fmt"
)

type DefenseType struct {
	raw string
}

func (s DefenseType) Complete() bool {
	return s.raw != ""
}

func (s DefenseType) String() string {
	return s.raw
}
func isValidDefenseType(value string) bool {
	validTypes := map[string]struct{}{
		"prompt-rewriting": {},
	}
	_, exists := validTypes[value]
	return exists
}

func NewDefenseType(value string) (DefenseType, error) {
	if value == "" {
		return DefenseType{}, errors.New("model cannot be empty")
	}
	if !isValidDefenseType(value) {
		return DefenseType{}, fmt.Errorf("invalid firewall type: %s", value)
	}
	return DefenseType{value}, nil
}
