package model

import (
	"netrunner/types"
	"time"
)

type Model struct {
	Name      types.Name   // User-provided name
	Model     types.Model  // Real model name to use with API
	ApiUrl    types.ApiUrl // URL to forward the request to
	CreatedAt time.Time
}
