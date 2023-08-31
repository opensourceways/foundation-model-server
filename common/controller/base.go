package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendBadRequestBody(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, newResponseCodeError(errorBadRequestBody, err))
}

func SendBadRequestParam(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, newResponseCodeError(errorBadRequestParam, err))
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
