package types

import "errors"

// ========================= MaxTokens =========================

type MaxTokens struct {
	value int
}

func (s MaxTokens) Complete() bool {
	return true
}

func (s MaxTokens) Int() int {
	return s.value
}

func isValidMaxTokens(value int) bool {
	if value <= 0 || value > 32000 {
		return false
	}
	return true
}

func NewMaxTokens(value int) (MaxTokens, error) {
	// Just validate - let API handle defaults
	if !isValidMaxTokens(value) {
		return MaxTokens{}, errors.New("invalid max_tokens value (must be > 0 and <= 32000)")
	}
	return MaxTokens{value}, nil
}

// ========================= Temperature =========================

type Temperature struct {
	value float32
}

func (s Temperature) Complete() bool {
	return true
}

func (s Temperature) Float32() float32 {
	return s.value
}

func isValidTemperature(value float32) bool {
	if value < 0 || value > 2 {
		return false
	}
	return true
}

func NewTemperature(value float32) (Temperature, error) {
	// Just validate - let API handle defaults
	if !isValidTemperature(value) {
		return Temperature{}, errors.New("invalid temperature value (must be between 0 and 2)")
	}
	return Temperature{value}, nil
}
