package customerrors

import "errors"

var (
	ErrFileTooLarge     = errors.New("file too large (max 5MB)")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrFileNotFound     = errors.New("file not found")
	ErrFileUploadFailed = errors.New("file upload failed")
	ErrNotImplemented   = errors.New("not implemented")
)
