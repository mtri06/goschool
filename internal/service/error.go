package service

type Error struct {
	msg     string
	errType string
	err     error
}

func NewError(msg, errType string, err error) *Error {
	return &Error{msg: msg, errType: errType, err: err}
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Type() string {
	return e.errType
}

func (e *Error) Msg() string {
	return e.msg
}

var (
	ErrValidationFailed = &Error{msg: "validation failed", errType: "validation_failed"}
	ErrNotFound         = &Error{msg: "resource not found", errType: "not_found"}
	ErrUnauthorized     = &Error{msg: "unauthorized", errType: "unauthorized"}
	ErrForbidden        = &Error{msg: "forbidden", errType: "forbidden"}
)
