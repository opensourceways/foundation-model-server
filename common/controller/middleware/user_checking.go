package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/foundation-model-server/allerror"
	commonstl "github.com/opensourceways/foundation-model-server/common/controller"
)

const headerPrivateToken = "PRIVATE-TOKEN"

var instance *accessTokenChecking

func Init(cfg *Config) {
	instance = &accessTokenChecking{
		accessToken: cfg.AccessToken,
	}
}

// Config
type Config struct {
	AccessToken string `json:"access_token" required:"true"`
}

// AccessTokenChecking
func AccessTokenChecking() gin.HandlerFunc {
	return instance.check
}

// accessTokenChecking
type accessTokenChecking struct {
	accessToken string
}

func (m *accessTokenChecking) check(ctx *gin.Context) {
	if err := m.doCheck(ctx); err != nil {
		commonstl.SendFailedResp(ctx, err)

		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func (m *accessTokenChecking) doCheck(ctx *gin.Context) error {
	t := m.token(ctx)
	if t == "" {
		logrus.Error("empty token forbidden")
		return allerror.New(allerror.ErrorCodeAccessTokenMissing, "")
	}

	if t != m.accessToken {
		logrus.Error("invalid token forbidden")
		return allerror.New(allerror.ErrorCodeAccessTokenInvalid, "")
	}

	return nil
}

func (m *accessTokenChecking) token(ctx *gin.Context) string {
	return ctx.GetHeader(headerPrivateToken)
}
