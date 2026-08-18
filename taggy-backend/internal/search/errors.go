package search

import "errors"

var (
	ErrInvalidQuery = errors.New("search query is invalid")
	ErrInvalidType  = errors.New("search type is invalid")
)
