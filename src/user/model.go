package user

import (
	"netrunner/src/types"
	"time"
)

type Model struct {
	Name      types.Name      // User-provided name
	Model     types.UserModel // Real model name to use with API
	ApiUrl    types.ApiUrl    // URL to forward the request to
	CreatedAt time.Time
}
