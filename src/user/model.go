package user

import (
	"covalence/src/types"
	"net/url"
	"time"
)

// ========================= Model =========================

type Model struct {
	Name      types.Name    // User-provided name
	Model     types.ModelID // Real model name to use with API
	APIURL    *url.URL
	CreatedAt time.Time
	Status    types.Status // Status of the model (active, inactive, etc.)
	Provider  types.ModelProvider
}
