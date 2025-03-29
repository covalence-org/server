package types

import (
	"errors"
	"fmt"
)

type FirewallType struct {
	raw string
}

func (s FirewallType) Complete() bool {
	return s.raw != ""
}

func (s FirewallType) String() string {
	return s.raw
}
func isValidFirewallType(value string) bool {
	validTypes := map[string]struct{}{
		"prompt-injection":   {},
		"malicious-intent":   {},
		"custom":             {},
		"policy-violation":   {},
		"sensitive-data":     {},
		"hallucination-risk": {},
		"spam":               {},
		"obfuscation":        {},
	}
	_, exists := validTypes[value]
	return exists
}

func NewFirewallType(value string) (FirewallType, error) {
	if value == "" {
		return FirewallType{}, errors.New("model cannot be empty")
	}
	if !isValidFirewallType(value) {
		return FirewallType{}, fmt.Errorf("invalid firewall type: %s", value)
	}
	return FirewallType{value}, nil
}
