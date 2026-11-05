package fs

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotDirectory    = errors.New("not a directory")
	ErrIsDirectory     = errors.New("is a directory")
	ErrDirectoryNotEmpty = errors.New("directory not empty")
	ErrInvalidName     = errors.New("invalid name")
	ErrPermission      = errors.New("permission denied")
	ErrSymlinkLoop     = errors.New("symlink loop")
	ErrWriterOpen      = errors.New("writer already open")
	ErrTypeMismatch    = errors.New("type mismatch")
	ErrNotChild        = errors.New("not a child of working directory")
	ErrCanceled        = errors.New("operation canceled")
)
