package allerror

const (
	ErrorCodeAccessTokenMissing = iota + 10000
	ErrorCodeAccessTokenInvalid
	ErrorCodeSensitiveContent
	ErrorCodeTooManyRequest
	ErrorCodeReqTimeout
	ErrorNotFound
	ErrorNotAllow
	ErrorInternalError
	ErrorFinetune
	ErrorBadRequest
	ErrorBadRequestParam
	ErrorBadRequestBody
	ErrorPermissionDeny
	ErrorSystemError
)

var errTable = map[int]string{
	ErrorCodeAccessTokenMissing: "access token missing",
	ErrorCodeAccessTokenInvalid: "access token invalid",
	ErrorCodeSensitiveContent:   "sensitive content",
	ErrorCodeTooManyRequest:     "too many requests",
	ErrorCodeReqTimeout:         "request timeout",
	ErrorNotFound:               "page not found",
	ErrorNotAllow:               "method not allowed",
	ErrorInternalError:          "internal error",
	ErrorFinetune:               "finetune error",
	ErrorBadRequest:             "bad_request",

	ErrorSystemError:     "system_error",
	ErrorBadRequestBody:  "bad_request_body",
	ErrorBadRequestParam: "bad_request_param",
	ErrorPermissionDeny:  "permission denied",
}

// errorImpl
type errorImpl struct {
	code int
	msg  string
}

func (e errorImpl) Error() string {
	return e.msg
}

func (e errorImpl) ErrorCode() int {
	return e.code
}

// New
func New(code int, msg string) errorImpl {
	v := errorImpl{
		code: code,
	}

	if msg == "" {
		v.msg = errTable[code]
	} else {
		v.msg = msg
	}

	return v
}

// notfoudError
type notfoudError struct {
	errorImpl
}

func (e notfoudError) NotFound() {}

// NewNotFound
func NewNotFound(msg string) notfoudError {
	return notfoudError{New(ErrorNotFound, msg)}
}
