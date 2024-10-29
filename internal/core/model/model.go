package model

import "errors"

var (
	ErrorBadRequest = errors.New("bad request")
	ErrorNotFound   = errors.New("not found")
)
