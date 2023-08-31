package qaimpl

type Config struct {
	MaxLengthOfQuestion  int    `json:"max_length_of_question"`
	ChatURL              string `json:"chat_url"                 required:"true"`
	ListModelsURL        string `json:"list_models_url"          required:"true"`
	GetWorkerAddressURL  string `json:"get_worker_address_url"   required:"true"`
	RefreshAllWorkersURL string `json:"refresh_all_workers_url"  required:"true"`
}

func (cfg *Config) SetDefault() {
	if cfg.MaxLengthOfQuestion <= 0 {
		cfg.MaxLengthOfQuestion = 1000
	}
}
