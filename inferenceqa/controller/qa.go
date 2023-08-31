package controller

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	commonctl "github.com/opensourceways/foundation-model-server/common/controller"
	"github.com/opensourceways/foundation-model-server/common/controller/middleware"
	"github.com/opensourceways/foundation-model-server/inferenceqa/app"
)

type QAController struct {
	service app.QAService
}

func AddRouteForQAController(r *gin.RouterGroup, s app.QAService) {
	ctl := QAController{
		service: s,
	}

	m := middleware.AccessTokenChecking()

	r.POST("/v1/inference/qa", m, ctl.Ask)
}

// Ask
// @Summary send a question
// @Description send a question
// @Tags  QA
// @Accept json
// @Param  param  body  qaRequest  true  "body of asking a question"
// @Success 201
// @Failure 400 {object} commonctl.ResponseData
// @Router /v1/inference/qa [post]
func (ctl QAController) Ask(ctx *gin.Context) {
	var req qaRequest

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
// @Tags  QA
// @Accept json
// @Success 200 {object} commonctl.ResponseData
// @Router /v1/inference/qa [get]
func (ctl QAController) Models(ctx *gin.Context) {
	commonctl.SendRespOfGet(ctx, ctl.service.Models())
}
