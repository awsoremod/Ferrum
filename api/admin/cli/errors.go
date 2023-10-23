package main

import "errors"

var (
	ErrRequired = errors.New("required value")
	ErrNotFound = errors.New("not found")
	ErrNil      = errors.New("nil data")
)
