package register

import (
	"covalence/src/types"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type rawModelProviders struct {
	Provider string   `yaml:"provider"`
	Models   []string `yaml:"models"`
	APIURL   string   `yaml:"api_url"`
}

type ModelProvider struct {
	Models   []types.ModelID
	Provider types.ModelProvider
	APIURL   types.APIURL
}

func ReadModelProviders() (*[]ModelProvider, error) {
	data, err := os.ReadFile("providers.yaml") // or whatever your file is named
	if err != nil {
		log.Fatalf("error reading file: %v", err)
		return &[]ModelProvider{}, err
	}

	var rawModelProviders []rawModelProviders
	err = yaml.Unmarshal(data, &rawModelProviders)
	if err != nil {
		log.Fatalf("error unmarshalling YAML: %v", err)
		return &[]ModelProvider{}, err
	}

	var modelProviders []ModelProvider
	for _, rawModelProvider := range rawModelProviders {
		provider, err := types.NewModelProvider(rawModelProvider.Provider)
		if err != nil {
			log.Fatalf("error creating provider: %v", err)
			continue
		}
		models := make([]types.ModelID, 0, len(rawModelProvider.Models))
		for _, model := range rawModelProvider.Models {
			modelID, err := types.NewModelID(model)
			if err != nil {
				log.Fatalf("error creating model ID: %v", err)
				continue
			}
			models = append(models, modelID)
		}
		apiURL, err := types.NewAPIURL(rawModelProvider.APIURL)
		if err != nil {
			log.Fatalf("error creating API URL: %v", err)
			continue
		}

		modelProviders = append(modelProviders, ModelProvider{
			Models:   models,
			Provider: provider,
			APIURL:   apiURL,
		})
	}

	return &modelProviders, nil
}
