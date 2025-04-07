package register

import (
	"covalence/src/user"
	"fmt"
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

// RegisterModel adds model information
func (r *Registry) Register(modelInfo user.Model) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	// check if model name already exists
	if _, exists := r.Models[modelInfo.Name.String()]; exists {
		return fmt.Errorf("model with name %s already exists", modelInfo.Name.String())
	}
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
