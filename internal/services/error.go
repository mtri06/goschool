package services

type Error struct {
	Msg  string
	Type string
	Err  error
}

func NewError(msg, errType string, err error) *Error {
	return &Error{Msg: msg, Type: errType, Err: err}
}

func (e *Error) Error() string {
	return e.Msg
}

func (e *Error) Unwrap() error {
	return e.Err
}

var (
	ErrValidationFailed = &Error{Msg: "validation failed", Type: "validation_failed", Err: nil}
	ErrNotFound         = &Error{Msg: "resource not found", Type: "not_found", Err: nil}
	ErrUnauthorized     = &Error{Msg: "unauthorized", Type: "unauthorized", Err: nil}
	ErrForbidden        = &Error{Msg: "forbidden", Type: "forbidden", Err: nil}
)
