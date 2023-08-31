package config

import (
	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/inferenceqa/infrastructure/moderationimpl"
	"github.com/opensourceways/foundation-model-server/inferenceqa/infrastructure/qaimpl"
)

func LoadConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := utils.LoadFromYaml(path, cfg); err != nil {
		return nil, err
	}

	cfg.SetDefault()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

type configValidate interface {
	Validate() error
}

type configSetDefault interface {
	SetDefault()
}

/*
type domainConfig struct {
	domain.Config

	DomainPrimitive dp.Config `json:"domain_primitive"`
}
*/

type inferenceQAConfig struct {
	MaxConcurrent int           `json:"max_concurrent"`
	Model         qaimpl.Config `json:"model"`
}

func (cfg *inferenceQAConfig) SetDefault() {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 100
	}
}

type Config struct {
	Middleware middleware.Config     `json:"middleware"`
	Moderation moderationimpl.Config `json:"moderation"`
	Chat       inferenceQAConfig     `json:"chat"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.Middleware,
		&cfg.Moderation,
		&cfg.Chat,
		&cfg.Chat.Model,
	}
}

func (cfg *Config) SetDefault() {
	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configSetDefault); ok {
			f.SetDefault()
		}
	}
}

func (cfg *Config) Validate() error {
	if _, err := utils.BuildRequestBody(cfg, ""); err != nil {
		return err
	}

	items := cfg.configItems()
	for _, i := range items {
		if f, ok := i.(configValidate); ok {
			if err := f.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}
