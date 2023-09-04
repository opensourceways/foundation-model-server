package config

import (
	"os"

	"github.com/opensourceways/server-common-lib/utils"

	"github.com/opensourceways/foundation-model-server/chat/infrastructure/chatadapter"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/common/infrastructure/moderationadapter"
)

func LoadConfig(path string) (Config, error) {
	cfg := Config{}

	if err := utils.LoadFromYaml(path, &cfg); err != nil {
		return cfg, err
	}

	if err := os.Remove(path); err != nil {
		return cfg, err
	}

	cfg.SetDefault()

	err := cfg.Validate()

	return cfg, err
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

type finetuneConfig struct {
	Kubeconfig string `json:"kubeconfig"`
	Namespace  string `json:"namespace"`
	Tokens     string `json:"token_file"`
}

func (cfg *chatConfig) SetDefault() {
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 100
	}
}

type Config struct {
	Chat       chatConfig               `json:"chat"`
	Middleware middleware.Config        `json:"middleware"`
	Moderation moderationadapter.Config `json:"moderation"`
	Finetune   finetuneConfig           `json:"finetune"`
}

func (cfg *Config) configItems() []interface{} {
	return []interface{}{
		&cfg.Chat,
		&cfg.Chat.Model,
		&cfg.Middleware,
		&cfg.Moderation,
		&cfg.Finetune,
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
