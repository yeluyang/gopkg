package errs

import "github.com/yeluyang/gopkg/errorx"

const (
	CodeAlreadyRunning errorx.Code = 1001
	CodeActivate       errorx.Code = 1002
	CodeSource         errorx.Code = 1003
	CodeAccept         errorx.Code = 1004
	CodeDrain          errorx.Code = 1005
	CodeCancelled      errorx.Code = 1006
	CodePanic          errorx.Code = 1007
)
