package errors_repo

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrZeroLength = errors.New("zero length")
)
