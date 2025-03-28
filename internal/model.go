package internal

import (
	"log"
	"netrunner/types"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Model struct {
	Model  types.InternalModel // Real model name to use with API
	ApiUrl types.ApiUrl        // URL to forward the request to
	Type   types.InternalModelType
}

var (
	models     []Model
	modelsOnce sync.Once
)

// LoadModels reads the models.yaml file, parses it into a list of Model structs, and saves it.
func LoadModels(filePath string) {
	modelsOnce.Do(func() {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}

		var rawModels []struct {
			Model  string `yaml:"model"`
			ApiUrl string `yaml:"api_url"`
			Type   string `yaml:"type"`
		}

		if err := yaml.Unmarshal(data, &rawModels); err != nil {
			log.Fatalf("failed to parse YAML: %v", err)
		}

		var parsedModels []Model
		for _, rawModel := range rawModels {
			model, err := types.NewInternalModel(rawModel.Model)
			if err != nil {
				log.Fatalf("failed to parse model: %v", err)
				continue
			}

			apiUrl, err := types.NewApiUrl(rawModel.ApiUrl)
			if err != nil {
				log.Fatalf("failed to parse API URL: %v", err)
				continue
			}

			modelType, err := types.NewInternalModelType(rawModel.Type)
			if err != nil {
				log.Fatalf("failed to parse model type: %v", err)
				continue
			}

			parsedModels = append(parsedModels, Model{
				Model:  model,
				ApiUrl: apiUrl,
				Type:   modelType,
			})
		}

		models = parsedModels
	})
}

func CheckModelExists(model types.InternalModel) bool {
	for _, m := range models {
		if m.Model == model {
			return true
		}
	}
	return false
}

// GetModels returns the loaded models.
func GetModels() []Model {
	return models
}
