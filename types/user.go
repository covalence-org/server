package types

import (
	"errors"
)

type UserModel struct {
	raw string
}

func (s UserModel) Complete() bool {
	return s.raw != ""
}

func (s UserModel) String() string {
	return s.raw
}

func NewUserModel(value string) (UserModel, error) {
	if value == "" {
		return UserModel{}, errors.New("Model cannot be empty")
	}
	return UserModel{value}, nil
}
