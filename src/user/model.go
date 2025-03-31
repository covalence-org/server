package user

import (
	"net/url"
	"netrunner/src/types"
	"time"
)

// ========================= Model =========================

type Model struct {
	Name      types.Name    // User-provided name
	Model     types.ModelID // Real model name to use with API
	APIURL    *url.URL
	CreatedAt time.Time
}
