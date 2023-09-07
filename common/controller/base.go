package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opensourceways/foundation-model-server/allerror"
)

func SendBadRequestBody(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, newResponseCodeError(fmt.Sprint(allerror.ErrorBadRequestBody), err))
}

func SendBadRequestParam(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, newResponseCodeError(fmt.Sprint(allerror.ErrorBadRequestParam), err))
}

// Post
func PostSuccessfully(ctx *gin.Context) {
	ctx.JSON(http.StatusCreated, newResponseCodeMsg("", "success"))
}

func SendRespOfPost(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusCreated, newResponseData(data))
}

// Put
func PutSuccessfully(ctx *gin.Context) {
	ctx.JSON(http.StatusAccepted, newResponseCodeMsg("", "success"))
}

// Get
func SendRespOfGet(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, newResponseData(data))
}

func SendFailedResp(ctx *gin.Context, err error) {
	sc, code := httpError(err)

	ctx.JSON(sc, newResponseCodeMsg(code, err.Error()))
}
