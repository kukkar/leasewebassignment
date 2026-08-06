package store

import (
	"errors"
	"fmt"
)

var (
	ErrRepositoryUninitialized = errors.New("repository uninitialized")
	ErrSourcePathRequired      = errors.New("source path required")
)

type StoreError struct {
	Op  string
	Err error
}

func (e *StoreError) Error() string {
	return fmt.Sprintf("store %s: %v", e.Op, e.Err)
}

func (e *StoreError) Unwrap() error {
	return e.Err
}
