package chatadapter

type configSetDefault interface {
	SetDefault()
}

var _ configSetDefault = (*Config)(nil)

type Config struct {
	chatConfig

	modelConfig
}

type chatConfig struct {
	//Timeout unit is second
	Timeout             int    `json:"timeout"`
	MaxLengthOfQuestion int    `json:"max_length_of_question"`
	ChatURL             string `json:"chat_url" required:"true"`
}

func (cfg *chatConfig) SetDefault() {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 180
	}

	if cfg.MaxLengthOfQuestion <= 0 {
		cfg.MaxLengthOfQuestion = 1000
	}
}

type modelConfig struct {
	ListModelsURL        string `json:"list_models_url"          required:"true"`
	GetWorkerAddressURL  string `json:"get_worker_address_url"   required:"true"`
	RefreshAllWorkersURL string `json:"refresh_all_workers_url"  required:"true"`
}
