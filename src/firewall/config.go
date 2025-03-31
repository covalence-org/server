package firewall

import (
	"fmt"
	"netrunner/src/internal"
	"netrunner/src/types"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Firewall struct {
	Enabled           bool
	ID                uuid.UUID
	Type              types.FirewallType
	Model             internal.Model
	BlockingThreshold float32
}

type Config struct {
	Name      string
	Firewalls []Firewall
}

type rawFirewall struct {
	Enabled           bool    `yaml:"enabled"`
	Type              string  `yaml:"type"`
	Model             string  `yaml:"model"`
	BlockingThreshold float32 `yaml:"blocking_threshold"`
}

type rawConfig struct {
	Name      string        `yaml:"name"`
	Firewalls []rawFirewall `yaml:"firewalls"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var raw rawConfig
	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	for _, rf := range raw.Firewalls {
		ft, err := types.NewFirewallType(rf.Type)
		if err != nil {
			return Config{}, fmt.Errorf("invalid firewall type: %w", err)
		}

		modelID, err := types.NewModelID(rf.Model)
		if err != nil {
			return Config{}, fmt.Errorf("invalid model: %w", err)
		}

		model, err := internal.GetModel(modelID)
		if err != nil {
			return Config{}, fmt.Errorf("failed to get model: %w", err)
		}

		cfg.Firewalls = append(cfg.Firewalls, Firewall{
			Enabled:           rf.Enabled,
			Type:              ft,
			Model:             model,
			BlockingThreshold: rf.BlockingThreshold,
		})
	}

	return cfg, nil
}
