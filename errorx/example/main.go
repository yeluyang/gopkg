package main

import (
	"errors"
	"fmt"

	"github.com/yeluyang/gopkg/errorx"
)

// Define error codes for your application
const (
	ErrNotFound    errorx.Code = 1001
	ErrUnauthorized errorx.Code = 1002
	ErrInternal    errorx.Code = 5000
)

func main() {
	// Example 1: Create error with message
	fmt.Println("=== Example 1: Create error with message ===")
	err := ErrNotFound.With("user not found")
	fmt.Printf("Error: %s\n", err)
	fmt.Printf("Code: %d\n", err.Code())

	// Example 2: Create error with formatted message
	fmt.Println("\n=== Example 2: Create error with formatted message ===")
	userID := "user123"
	err = ErrNotFound.Format("user %q does not exist", userID)
	fmt.Printf("Error: %s\n", err)

	// Example 3: Wrap existing error
	fmt.Println("\n=== Example 3: Wrap existing error ===")
	dbErr := errors.New("connection refused")
	err = ErrInternal.From(dbErr)
	fmt.Printf("Error: %s\n", err)
	fmt.Printf("Unwrapped: %v\n", errors.Unwrap(err))

	// Example 4: Check error code using errors.Is
	fmt.Println("\n=== Example 4: Check error code using errors.Is ===")
	if errors.Is(err, ErrInternal.With("any message")) {
		fmt.Println("Error is an internal error")
	}

	// Example 5: Print with stack trace
	fmt.Println("\n=== Example 5: Print with stack trace ===")
	err = createError()
	fmt.Printf("%v\n", err)

	// Example 6: Use errors.As with wrapped errors
	fmt.Println("\n=== Example 6: Use errors.As with wrapped errors ===")
	customErr := &ValidationError{Field: "email", Message: "invalid format"}
	err = ErrUnauthorized.From(customErr)
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("Validation error on field %q: %s\n", valErr.Field, valErr.Message)
	}
}

func createError() *errorx.Error {
	return innerFunc()
}

func innerFunc() *errorx.Error {
	return ErrNotFound.With("resource not found")
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
