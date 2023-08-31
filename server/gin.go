package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opensourceways/server-common-lib/interrupts"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/config"
	"github.com/opensourceways/foundation-model-server/docs"
	qaapp "github.com/opensourceways/foundation-model-server/inferenceqa/app"
	qactl "github.com/opensourceways/foundation-model-server/inferenceqa/controller"
	"github.com/opensourceways/foundation-model-server/inferenceqa/domain/dp"
	chatservice "github.com/opensourceways/foundation-model-server/inferenceqa/domain/service"
	"github.com/opensourceways/foundation-model-server/inferenceqa/infrastructure/flowcontrollerimpl"
	"github.com/opensourceways/foundation-model-server/inferenceqa/infrastructure/moderationimpl"
	"github.com/opensourceways/foundation-model-server/inferenceqa/infrastructure/qaimpl"
)

func StartWebServer(port int, timeout time.Duration, cfg *config.Config) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logRequest())

	setRouter(r, cfg)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	defer interrupts.WaitForGracefulShutdown()

	interrupts.ListenAndServe(srv, timeout)
}

//setRouter init router
func setRouter(engine *gin.Engine, cfg *config.Config) {
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Title = "Foundation Model"
	docs.SwaggerInfo.Description = "set header: 'PRIVATE-TOKEN=xxx'"

	v1 := engine.Group(docs.SwaggerInfo.BasePath)
	setApiV1(v1, cfg)

	engine.UseRawPath = true
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func setApiV1(v1 *gin.RouterGroup, cfg *config.Config) {
	m := moderationimpl.Init(&cfg.Moderation)

	chat := qaimpl.ChatServiceInstance()

	s := chatservice.NewQAService(
		m, flowcontrollerimpl.Init(cfg.Chat.MaxConcurrent), chat,
	)

	dp.Init(cfg.Chat.Model.MaxLengthOfQuestion, chat)

	middleware.Init(&cfg.Middleware)

	qactl.AddRouteForQAController(
		v1, qaapp.NewQAService(s),
	)
}

func logRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		endTime := time.Now()

		logrus.Infof(
			"| %d | %d | %s | %s |",
			c.Writer.Status(),
			endTime.Sub(startTime),
			c.Request.Method,
			c.Request.RequestURI,
		)
	}
}
