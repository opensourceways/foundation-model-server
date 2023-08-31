package moderationimpl

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	moderation "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/moderation/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/moderation/v3/model"

	"github.com/opensourceways/foundation-model-server/allerror"
)

func Init(cfg *Config) *moderationImpl {
	auth := basic.NewCredentialsBuilder().
		WithAk(cfg.AccessKey).
		WithSk(cfg.SecretKey).
		WithIamEndpointOverride(cfg.IAMEndpint).
		Build()

	cli := moderation.NewModerationClient(
		moderation.ModerationClientBuilder().
			WithRegion(region.NewRegion(cfg.Region, cfg.Endpoint)).
			WithCredential(auth).
			Build(),
	)

	return &moderationImpl{cli}
}

type moderationImpl struct {
	cli *moderation.ModerationClient
}

func (s *moderationImpl) CheckText(content string) error {
	request := &model.RunTextModerationRequest{
		Body: &model.TextDetectionReq{
			Data: &model.TextDetectionDataReq{
				Text: content,
			},
			EventType: "comment",
		},
	}

	resp, err := s.cli.RunTextModeration(request)
	if err != nil {
		return err
	}

	if *resp.Result.Suggestion != "pass" {
		return allerror.New(allerror.ErrorCodeSensitiveContent, "")
	}

	return nil
}
