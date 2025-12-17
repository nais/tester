package reporter

import (
	"context"
	"fmt"
	"time"
)

// InfoArg represents a single argument with optional name
type InfoArg struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value"`
}

// Error represents a structured error with optional expected/actual values for diffs
type Error struct {
	Message  string `json:"message"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

// NewError creates a simple error with just a message
func NewError(msg string, args ...any) *Error {
	return &Error{
		Message: fmt.Sprintf(msg, args...),
	}
}

// NewDiffError creates an error with expected and actual values for diff display
func NewDiffError(diff string, expected, actual any) *Error {
	return &Error{
		Message:  fmt.Sprintf("diff -want +got:\n%v", diff),
		Expected: expected,
		Actual:   actual,
	}
}

// InfoType represents the type of information being reported
type InfoType string

const (
	// InfoTypeHelper is used when a helper function is called
	InfoTypeHelper InfoType = "helper"
	// InfoTypeRequest is used for HTTP/GraphQL requests
	InfoTypeRequest InfoType = "request"
	// InfoTypeResponse is used for HTTP/GraphQL responses
	InfoTypeResponse InfoType = "response"
	// InfoTypeQuery is used for SQL/GraphQL queries
	InfoTypeQuery InfoType = "query"
	// InfoTypeResult is used for query results
	InfoTypeResult InfoType = "result"
)

// Info represents a piece of information about a test execution
type Info struct {
	Type      InfoType      `json:"type"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Args      []InfoArg     `json:"args,omitempty"`
	Timestamp time.Duration `json:"timestamp"`
	Language  string        `json:"language,omitempty"`
}

type Reporter interface {
	RunFile(ctx context.Context, filename string, fn func(Reporter))
	RunTest(ctx context.Context, runner, name string, fn func(Reporter))
	ReportError(err *Error)
	Info(info Info)
}
