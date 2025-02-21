package repositories

import "errors"

// ErrNotFound используется, когда сущность не найдена в хранилище.
var ErrNotFound = errors.New("not found")
