package types

import (
	"errors"
)

// ======== Internal Model Types ==========

type Status struct {
	raw string
}

func (s Status) Complete() bool {
	return s.raw != ""
}

func (s Status) String() string {
	return s.raw
}

func Active() Status {
	return Status{"active"}
}

func Inactive() Status {
	return Status{"inactive"}
}

func isValidStatus(value string) bool {
	return value == "active" || value == "inactive"
}

func NewStatus(value string) (Status, error) {
	if value == "" {
		return Status{}, errors.New("status cannot be empty")
	}
	if !isValidStatus(value) {
		return Status{}, errors.New("invalid status type")
	}
	return Status{value}, nil
}
