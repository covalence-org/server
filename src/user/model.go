package user

import (
	"netrunner/src/types"
	"time"
)

// ========================= Model =========================

type Model struct {
	Name      types.Name    // User-provided name
	Model     types.ModelID // Real model name to use with API
	APIURL    types.APIURL  // URL to forward the request to
	CreatedAt time.Time
}
