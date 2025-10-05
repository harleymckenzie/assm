package apperror

import (
	"errors"
	"fmt"
	"os"
)

const (
	// Exit codes loosely based off of BSD sysexits.h conventions - https://man7.org/linux/man-pages/man3/sysexits.h.3head.html
	CodeOK               = 0  // EX_OK - Success
	CodeGeneralError     = 1  // EX_GENERAL - General error
	CodeInvalidUsage     = 2  // EX_USAGE - Misuse of shell builtins
	CodePluginNotFound   = 64 // EX_USAGE - Session manager plugin not found (usage error)
	CodeServiceError     = 69 // EX_UNAVAILABLE - Service unavailable (AWS service)
	CodeAuthError        = 77 // EX_NOPERM - Permission denied (AWS auth)
	CodeConfigError      = 78 // EX_CONFIG - Configuration error
	CodeNoInstancesFound = 79 // Custom - No instances found (app-specific)
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

// New creates a new Error with the specified exit code and error.
func New(code int, err error) *Error {
	return &Error{
		Code: code,
		Err:  err,
	}
}

// Exit prints the error message and exits with the appropriate code.
func Exit(err error) {
	if err == nil {
		os.Exit(CodeOK)
	}
	fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	os.Exit(Code(err))
}

// Code returns the intended exit code for err (default 1 if not tagged).
func Code(err error) int {
	var ee *Error
	if errors.As(err, &ee) {
		return ee.Code
	}
	return CodeGeneralError
}
