package register

import (
	"netrunner/model"
	"sync"
)

// ModelRegistry stores registered models
type Registry struct {
	Mu     sync.RWMutex
	Models map[string]model.Model
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *Registry {
	return &Registry{
		Models: make(map[string]model.Model),
	}
}

// RegisterModel adds or updates model information
func (r *Registry) Register(modelInfo model.Model) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.Models[modelInfo.Name.String()] = modelInfo

	return nil
}

// GetModelInfo retrieves model information by custom name
func (r *Registry) GetInfo(name string) (model.Model, bool) {
	r.Mu.RLock()
	defer r.Mu.RUnlock()
	info, exists := r.Models[name]
	return info, exists
}
