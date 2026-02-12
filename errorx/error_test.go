package errorx

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	CodeNotFound    Code = 404
	CodeServerError Code = 500
	CodeBadRequest  Code = 400
)

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(suiteError))
}

type suiteError struct {
	suite.Suite
}

func (s *suiteError) TestError() {
	err := CodeNotFound.With("resource not found")
	s.Equal("[404] resource not found", err.Error())
}

func (s *suiteError) TestNil() {
	var err *Error
	s.Equal("<nil>", err.Error())
}

func (s *suiteError) TestNilErr() {
	err := &Error{code: 42}
	s.Equal("<nil>", err.Error())
	s.Equal(Code(0), err.Code())
	s.Nil(err.Unwrap())
	s.False(err.Is(CodeNotFound.With("x")))
	s.Equal("<nil>", fmt.Sprintf("%v", err))
}

func (s *suiteError) TestCode() {
	err := CodeServerError.With("internal error")
	s.Equal(CodeServerError, err.Code())
}

func (s *suiteError) TestCode_Nil() {
	var err *Error
	s.Equal(Code(0), err.Code())
}

func (s *suiteError) TestUnwrap() {
	inner := errors.New("inner error")
	err := CodeBadRequest.From(inner)

	s.Equal(inner, err.Unwrap())
	s.Equal(inner, errors.Unwrap(err))
}

func (s *suiteError) TestUnwrap_Nil() {
	var err *Error
	s.Nil(err.Unwrap())
}

func (s *suiteError) TestIs_SameCode() {
	err1 := CodeNotFound.With("not found 1")
	err2 := CodeNotFound.With("not found 2")

	s.True(errors.Is(err1, err2))
}

func (s *suiteError) TestIs_DifferentCode() {
	err1 := CodeNotFound.With("not found")
	err2 := CodeServerError.With("server error")

	s.False(errors.Is(err1, err2))
}

func (s *suiteError) TestIs_NonErrorType() {
	err := CodeNotFound.With("not found")
	s.False(errors.Is(err, errors.New("plain error")))
}

func (s *suiteError) TestIs_NilReceiver() {
	var nilErr *Error
	err := CodeNotFound.With("not found")

	s.False(nilErr.Is(err))
	s.False(nilErr.Is(nil))
}

func (s *suiteError) TestFrom_Errorx() {
	original := CodeNotFound.With("not found")
	e, ok := From(original)
	s.True(ok)
	s.Equal(original, e)
}

func (s *suiteError) TestFrom_WrappedErrorx() {
	original := CodeNotFound.With("not found")
	wrapped := fmt.Errorf("wrap: %w", original)
	e, ok := From(wrapped)
	s.True(ok)
	s.Equal(original, e)
}

func (s *suiteError) TestFrom_PlainError() {
	e, ok := From(errors.New("plain"))
	s.False(ok)
	s.Nil(e)
}

func (s *suiteError) TestFrom_Nil() {
	e, ok := From(nil)
	s.False(ok)
	s.Nil(e)
}

func (s *suiteError) TestMustFrom_Errorx() {
	original := CodeNotFound.With("not found")
	e := MustFrom(original)
	s.Equal(original, e)
}

func (s *suiteError) TestMustFrom_PlainError() {
	plain := errors.New("plain")
	e := MustFrom(plain)
	s.NotNil(e)
	s.Equal(CodeUnknown, e.Code())
	s.Equal(plain, e.Unwrap())
}

func (s *suiteError) TestMustFrom_Nil() {
	e := MustFrom(nil)
	s.NotNil(e)
	s.Equal("<nil>", e.Error())
	s.Equal(CodeOK, e.Code())
	s.Nil(e.Unwrap())
}

func (s *suiteError) TestAs() {
	inner := &customError{msg: "custom"}
	err := CodeBadRequest.From(inner)

	var target *customError
	s.True(errors.As(err, &target))
	s.Equal("custom", target.msg)
}

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

type CodeSuite struct {
	suite.Suite
}

func TestCodeSuite(t *testing.T) {
	suite.Run(t, new(CodeSuite))
}

func (s *CodeSuite) TestFormat() {
	err := CodeNotFound.Format("user %s not found", "alice")
	s.Equal("[404] user alice not found", err.Error())
}

func (s *CodeSuite) TestWith() {
	err := CodeServerError.With("internal server error")
	s.Equal("[500] internal server error", err.Error())
}

func (s *CodeSuite) TestFrom() {
	inner := errors.New("database connection failed")
	err := CodeServerError.From(inner)

	s.Equal(CodeServerError, err.Code())
	s.Equal(inner, err.Unwrap())
}

type FormatSuite struct {
	suite.Suite
}

func TestFormatSuite(t *testing.T) {
	suite.Run(t, new(FormatSuite))
}

func (s *FormatSuite) TestFormat_S() {
	err := CodeNotFound.With("not found")
	s.Equal("[404] not found", fmt.Sprintf("%s", err))
}

func (s *FormatSuite) TestFormat_Q() {
	err := CodeNotFound.With("not found")
	s.Equal(`"[404] not found"`, fmt.Sprintf("%q", err))
}

func (s *FormatSuite) TestFormat_V_ContainsMessage() {
	err := CodeNotFound.With("not found")
	v := fmt.Sprintf("%v", err)

	s.Contains(v, "[404] not found")
}

func (s *FormatSuite) TestFormat_V_ContainsStackTrace() {
	err := CodeNotFound.With("not found")
	v := fmt.Sprintf("%v", err)

	s.Contains(v, "Stack Trace:")
	s.Contains(v, "TestFormat_V_ContainsStackTrace")
}

func (s *FormatSuite) TestFormat_Nil() {
	var err *Error
	s.Equal("<nil>", fmt.Sprintf("%v", err))
}

func (s *FormatSuite) TestFormat_V_StackTraceOrder() {
	err := createNestedError()
	v := fmt.Sprintf("%v", err)

	innerIdx := strings.Index(v, "innerFunc")
	outerIdx := strings.Index(v, "createNestedError")

	s.Greater(outerIdx, innerIdx, "inner function should appear before outer in stack trace")
}

func createNestedError() *Error {
	return innerFunc()
}

func innerFunc() *Error {
	return CodeNotFound.With("nested error")
}

type NewSuite struct {
	suite.Suite
}

func TestNewSuite(t *testing.T) {
	suite.Run(t, new(NewSuite))
}

func (s *NewSuite) TestNew() {
	inner := errors.New("inner")
	stack := []uintptr{1, 2, 3}
	err := New(CodeNotFound, inner, stack)

	s.Equal(CodeNotFound, err.Code())
	s.Equal(inner, err.Unwrap())
}
