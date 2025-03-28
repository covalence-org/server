package register

import (
	"netrunner/user"
	"sync"
)

// ModelRegistry stores registered models
type Registry struct {
	Mu     sync.RWMutex
	Models map[string]user.Model
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *Registry {
	return &Registry{
		Models: make(map[string]user.Model),
	}
}

// RegisterModel adds or updates model information
func (r *Registry) Register(modelInfo user.Model) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Models[modelInfo.Name.String()] = modelInfo

	return nil
}

// GetModelInfo retrieves model information by custom name
func (r *Registry) GetInfo(name string) (user.Model, bool) {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	info, exists := r.Models[name]
	return info, exists
}
