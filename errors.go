package sargs

import "errors"

var (
	ErrNotPtrToStruct       = errors.New("not a pointer to a struct")
	ErrUnsupportedFieldType = errors.New("unsupported file type")
	ErrInvalidDefaultValue  = errors.New("invalid default value")
	ErrInvalidArgPos        = errors.New("invalid argument postion")
	ErrMissingFlag          = errors.New("missing flag")
	ErrNotEnoughArgs        = errors.New("not enough arguments")
)
