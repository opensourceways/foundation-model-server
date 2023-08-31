package controller

import (
	"net/http"

	"github.com/opensourceways/foundation-model-server/allerror"
)

const (
	errorBadRequest      = "bad_request"
	errorSystemError     = "system_error"
	errorBadRequestBody  = "bad_request_body"
	errorBadRequestParam = "bad_request_param"
)

type errorCode interface {
	ErrorCode() string
}

type errorNotFound interface {
	errorCode

	NotFound()
}

func httpError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	sc := http.StatusInternalServerError
	code := errorSystemError

	if v, ok := err.(errorCode); ok {
		code = v.ErrorCode()

		if _, ok := err.(errorNotFound); ok {
			sc = http.StatusNotFound
		} else {
			switch code {
			case allerror.ErrorCodeAccessTokenInvalid:
				sc = http.StatusUnauthorized

			default:
				sc = http.StatusBadRequest
			}
		}
	}

	return sc, code
}
