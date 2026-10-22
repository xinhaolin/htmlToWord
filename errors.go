package htmlword

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidConfig       = errors.New("invalid converter configuration")
	ErrInvalidInput        = errors.New("invalid conversion input")
	ErrRuntimeUnavailable  = errors.New("Chromium runtime unavailable")
	ErrUnsupportedPlatform = errors.New("unsupported runtime platform")
	ErrClosed              = errors.New("converter is closed")
	ErrLimit               = errors.New("conversion limit exceeded")
	ErrResource            = errors.New("resource could not be loaded")
	ErrConversion          = errors.New("conversion failed")
	ErrStrict              = errors.New("strict mode rejected conversion diagnostics")
	ErrOutput              = errors.New("output write failed")
)

// Error identifies the failed stage and category. Use errors.Is for categories
// and context cancellation; errors.As exposes Stage and the underlying Cause.
type Error struct {
	Stage Stage
	Kind  error
	Cause error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("htmlword %s: %v", e.Stage, e.Kind)
	}
	return fmt.Sprintf("htmlword %s: %v: %v", e.Stage, e.Kind, e.Cause)
}
func (e *Error) Unwrap() error        { return e.Cause }
func (e *Error) Is(target error) bool { return target == e.Kind }
func failure(stage Stage, kind, cause error) error {
	return &Error{Stage: stage, Kind: kind, Cause: cause}
}
