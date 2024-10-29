package model

import "errors"

var (
	ErrorBadRequest = errors.New("bad request")
	ErrorNotFound   = errors.New("not found")
)

type Metric struct {
	Name  string
	Value float64
	Type  string
}
