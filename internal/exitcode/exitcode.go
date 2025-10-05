package exitcode

import (
	"errors"
	"fmt"
	"os"
)

const (
	// Exit codes loosely based off of BSD sysexits.h conventions - https://man7.org/linux/man-pages/man3/sysexits.h.3head.html
	CodeOK                 = 0  // EX_OK - Success
	CodeGeneralError       = 1  // EX_GENERAL - General error
	CodeInvalidUsage       = 2  // EX_USAGE - Misuse of shell builtins
	CodeUnknownCommand     = 64 // EX_USAGE - Unknown/unsupported command (usage error)
	CodeUnsupportedType    = 65 // EX_DATAERR - Unsupported type (data error)
	CodeNoDocumentsFound   = 66 // EX_NOINPUT - No documents match filters
	CodeServiceError       = 69 // EX_UNAVAILABLE - Service unavailable (AWS service)
	CodeAuthError          = 77 // EX_NOPERM - Permission denied (AWS auth)
	CodeConfigError        = 78 // EX_CONFIG - Configuration error
	CodeNoInstancesMatched = 79 // Custom - No instances matched (app-specific)
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func New(code int, err error) {
	fmt.Printf("[err] %s\n", err.Error())
	os.Exit(code)
}

// Code returns the intended exit code for err (default 1 if not tagged).
func Code(err error) int {
	var ee *Error
	if errors.As(err, &ee) {
		return ee.Code
	}
	return CodeGeneralError
}
