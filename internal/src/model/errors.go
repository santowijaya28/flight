package model

import "errors"

var (
	ErrAllProviderDown error = errors.New("all provider down")
	ErrInvalidInput    error = errors.New("invalid input")
)
