package ckjson

import "fmt"

type ReadError struct {
	errorMessage string
}

func NewReadError(errorMessage string) *ReadError {
	return &ReadError{errorMessage: fmt.Sprintf("could not read rune: %s", errorMessage)}
}

func (r *ReadError) Error() string { return r.errorMessage }

type SyntaxError struct {
	errorMessage string
}

func NewSyntaxError(errorMessage string) *SyntaxError {
	return &SyntaxError{errorMessage: errorMessage}
}

func (r *SyntaxError) Error() string { return r.errorMessage }
