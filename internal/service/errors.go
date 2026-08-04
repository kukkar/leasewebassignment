package service

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidUpload = errors.New("invalid upload payload")
)

type ServiceError struct {
	Op  string
	Err error
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("service %s: %v", e.Op, e.Err)
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

type CSVParseError struct {
	Row    int
	Column string
	Reason string
	Err    error
}

func (e *CSVParseError) Error() string {
	if e.Row > 0 {
		return fmt.Sprintf("csv parse row=%d column=%s reason=%s: %v", e.Row, e.Column, e.Reason, e.Err)
	}
	return fmt.Sprintf("csv parse column=%s reason=%s: %v", e.Column, e.Reason, e.Err)
}

func (e *CSVParseError) Unwrap() error {
	return e.Err
}
