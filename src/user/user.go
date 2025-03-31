package user

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	APIKeyID uuid.UUID
}

func GetUserByAPIKey(apiKey string) (User, error) {
	// hash api key

	// lookup api key id in db

	// for now, leave it random
	return User{
		ID:       uuid.New(),
		APIKeyID: uuid.New(),
	}, nil
}
