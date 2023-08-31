package config

import (
	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/foundation-model-server/chat/infrastructure/chatadapter"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/common/infrastructure/moderationadapter"
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

type chatConfig struct {
	MaxConcurrent int                `json:"max_concurrent"`
	Model         chatadapter.Config `json:"model"`
}

func (cfg *chatConfig) SetDefault() {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 100
	}
}

type Config struct {
	Middleware middleware.Config        `json:"middleware"`
	Moderation moderationadapter.Config `json:"moderation"`
	Chat       chatConfig               `json:"chat"`
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
