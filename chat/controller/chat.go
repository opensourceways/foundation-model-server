package controller

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/opensourceways/foundation-model-server/chat/app"
	commonctl "github.com/opensourceways/foundation-model-server/common/controller"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
)

type ChatController struct {
	service app.ChatAppService
}

func AddRouteForChatController(r *gin.RouterGroup, s app.ChatAppService) {
	ctl := ChatController{
		service: s,
	}

	m := middleware.AccessTokenChecking()

	r.POST("/v1/chat", m, ctl.Ask)
	r.POST("/v1/chat/models", m, ctl.Models)
}

// Ask
// @Summary ask a question
// @Description ask a question
// @Tags  Chat
// @Accept json
// @Param  param  body  askQuestionRequest  true  "body of asking a question"
// @Success 201
// @Failure 400 {object} commonctl.ResponseData
// @Router /v1/chat [post]
func (ctl ChatController) Ask(ctx *gin.Context) {
	var req askQuestionRequest

	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		commonctl.SendBadRequestBody(ctx, err)

		return
	}

	cmd, err := req.toCmd()
	if err != nil {
		commonctl.SendBadRequestParam(ctx, err)

		return
	}

	cmd.SteamWrite = func(doOnce func(io.Writer) (bool, error)) {
		ctx.Stream(func(w io.Writer) bool {
			done, err := doOnce(w)

			return !done && err == nil
		})
	}

	if err := ctl.service.Ask(&cmd); err != nil {
		commonctl.SendFailedResp(ctx, err)
	}
}

// Models
// @Summary list all models
// @Description list all models
// @Tags  Chat
// @Accept json
// @Success 200 {object} commonctl.ResponseData
// @Router /v1/chat/models [get]
func (ctl ChatController) Models(ctx *gin.Context) {
	commonctl.SendRespOfGet(ctx, ctl.service.Models())
}
