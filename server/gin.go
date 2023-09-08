package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/opensourceways/foundation-model-server/allerror"
	chatapp "github.com/opensourceways/foundation-model-server/chat/app"
	chatctl "github.com/opensourceways/foundation-model-server/chat/controller"
	"github.com/opensourceways/foundation-model-server/chat/domain/dp"
	chatservice "github.com/opensourceways/foundation-model-server/chat/domain/service"
	"github.com/opensourceways/foundation-model-server/chat/infrastructure/chatadapter"
	commonctl "github.com/opensourceways/foundation-model-server/common/controller"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/common/infrastructure/flowcontrolleradapter"
	"github.com/opensourceways/foundation-model-server/common/infrastructure/moderationadapter"
	"github.com/opensourceways/foundation-model-server/config"
	"github.com/opensourceways/foundation-model-server/docs"
	finetunectl "github.com/opensourceways/foundation-model-server/finetune/controller"
	"github.com/opensourceways/server-common-lib/interrupts"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func StartWebServer(port int, timeout time.Duration, cfg *config.Config) {
	r := gin.New()
	r.Use(timeoutMiddleware(cfg.Middleware.Timeout))
	r.Use(gin.Recovery())
	r.Use(logRequest())

	setRouter(r, cfg)

	r.NoRoute(func(c *gin.Context) {
		commonctl.SendFailedResp(c, allerror.New(allerror.ErrorNotFound, ""))
	})

	r.NoMethod(func(c *gin.Context) {
		commonctl.SendFailedResp(c, allerror.New(allerror.ErrorNotAllow, ""))
	})

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
	m := moderationadapter.Init(&cfg.Moderation)

	chat := chatadapter.ChatAdapter()

	s := chatservice.NewChatService(
		m, flowcontrolleradapter.Init(cfg.Chat.MaxConcurrent), chat,
	)

	dp.Init(cfg.Chat.Model.MaxLengthOfQuestion, chat)

	middleware.Init(&cfg.Middleware)

	chatctl.AddRouteForChatController(
		v1, chatapp.NewChatAppService(s),
	)
	finetunectl.RegisterRoutes(v1)
}

func timeoutResponse(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
		"code": fmt.Sprint(allerror.ErrorCodeReqTimeout),
		"msg":  http.StatusRequestTimeout,
	})
}

// RequestTimeout 处理请求超时
func timeoutMiddleware(t int) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(time.Duration(t)*time.Second),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
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
