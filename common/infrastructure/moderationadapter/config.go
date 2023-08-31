package moderationadapter

type Config struct {
	Endpoint   string `json:"endpoint"       required:"true"`
	AccessKey  string `json:"access_key"     required:"true"`
	SecretKey  string `json:"secret_key"     required:"true"`
	IAMEndpint string `json:"iam_endpoint"   required:"true"`
	Region     string `json:"region"         required:"true"`
}
