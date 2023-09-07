package controller

import (
	"fmt"
	"net/http"

	"github.com/opensourceways/foundation-model-server/allerror"
)

type errorCode interface {
	ErrorCode() int
}

type errorNotFound interface {
	errorCode

	NotFound()
}

var errTable = map[int]int{
	allerror.ErrorCodeAccessTokenMissing: http.StatusUnauthorized,
	allerror.ErrorCodeAccessTokenInvalid: http.StatusUnauthorized,
	allerror.ErrorCodeSensitiveContent:   http.StatusBadRequest,
	allerror.ErrorCodeTooManyRequest:     http.StatusTooManyRequests,
	allerror.ErrorCodeReqTimeout:         http.StatusRequestTimeout,
	allerror.ErrorNotFound:               http.StatusNotFound,
	allerror.ErrorInternalError:          http.StatusInternalServerError,
}

func httpError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "0"
	}

	sc := http.StatusInternalServerError
	code := allerror.ErrorInternalError

	if v, ok := err.(errorCode); ok {
		code = v.ErrorCode()

		if status, ok := errTable[code]; ok {
			sc = status
		} else {
			sc = http.StatusBadRequest
		}
	}

	return sc, fmt.Sprint(code)
}
