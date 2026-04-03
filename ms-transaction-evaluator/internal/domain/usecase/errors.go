package usecase

import "errors"

var ErrEventPublishFailed = errors.New("failed to publish transaction event")

var ErrTransactionNotFound = errors.New("transaction not found")

var ErrInvalidLimit = errors.New("invalid limit: must be between 1 and 100")

var ErrInvalidCursor = errors.New("invalid cursor")
