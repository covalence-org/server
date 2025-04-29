package dynamicdefense

import (
	"covalence/src/types"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name     string
	Defenses []Defense
}

type rawDefense struct {
	ID      string `yaml:"id"`
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"`
}

type rawConfig struct {
	Name      string       `yaml:"name"`
	Firewalls []rawDefense `yaml:"defenses"`
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
		dt, err := types.NewDefenseType(rf.Type)
		if err != nil {
			return Config{}, fmt.Errorf("invalid defense type: %w", err)
		}

		id, err := uuid.Parse(rf.ID)
		if err != nil {
			return Config{}, fmt.Errorf("invalid defense ID: %w", err)
		}

		cfg.Defenses = append(cfg.Defenses, Defense{
			Enabled: rf.Enabled,
			ID:      id,
			Type:    dt,
		})
	}

	return cfg, nil
}
