package errors

import "bytes"

type BizError interface {
	error
	GetCode() string
	GetMsg() string
	GetErrors() *[]BizError
}

type SimpleBizError struct {
	Code   string      `json:"code,omitempty"`
	Msg    string      `json:"msg,omitempty"`
	Errors *[]BizError `json:"errors,omitempty"`
}

type FieldError struct {
	*SimpleBizError
	Name string `json:"name,omitempty"`
}

func (err *SimpleBizError) Error() string {
	if err == nil {
		return ""
	}

	if err.Errors != nil {
		sb := bytes.Buffer{}
		sb.WriteString(err.Msg)

		for _, fe := range *err.Errors {
			sb.WriteByte('\n')
			sb.WriteString(fe.Error())
		}

		return sb.String()
	} else {
		return err.Msg
	}
}

func (err *SimpleBizError) GetCode() string {
	if err == nil {
		return ""
	}

	return err.Code
}

func (err *SimpleBizError) GetMsg() string {
	if err == nil {
		return ""
	}
	return err.Msg
}

func (err *SimpleBizError) AddError(cerr BizError) *SimpleBizError {
	if cerr == nil {
		return err
	}

	if err.Errors == nil {
		err.Errors = new([]BizError)
	}

	*err.Errors = append(*err.Errors, cerr)
	return err
}

func (err *SimpleBizError) GetErrors() *[]BizError {
	return err.Errors
}

func (err *SimpleBizError) HasError() bool {
	return err.Errors != nil && len(*err.Errors) > 0
}

const (
	Common_NotFound    = "c.NOT_FOUND"
	Common_RPCError    = "c.RPC_ERROR"
	Common_ServerError = "c.SERVER_ERROR"

	Common_InvalidParams = "c.INVALID_PARAMS"
	FIELD_BAD_FORMAT     = "BAD_FORMAT"
	Common_Unauthorized  = "c.UNAUTHORIZED"
	Common_Forbidden     = "c.FORBIDDEN"
	//Common_InvalidField  = "c.INVALID_FIELD"
)

func Empty() *SimpleBizError {
	return &SimpleBizError{}
}

func RPCFailed() *SimpleBizError {
	return &SimpleBizError{Code: Common_RPCError}
}

func RPCFailedWithMsg(msg string) *SimpleBizError {
	return &SimpleBizError{Code: Common_RPCError, Msg: msg}
}

func NotFound() *SimpleBizError {
	return &SimpleBizError{Code: Common_NotFound}
}

func Unauthorized() *SimpleBizError {
	return &SimpleBizError{Code: Common_Unauthorized}
}

func Forbidden() *SimpleBizError {
	return &SimpleBizError{Code: Common_Forbidden}
}

func NotFoundWithMsg(msg string) *SimpleBizError {
	return &SimpleBizError{Code: Common_NotFound, Msg: msg}
}

func ServerError() *SimpleBizError {
	return &SimpleBizError{Code: Common_ServerError}
}

func ServerErrorWithMsg(msg string) *SimpleBizError {
	return &SimpleBizError{Code: Common_ServerError, Msg: msg}
}

//func NotFoundWithMsg(msg string) BizError {
//	return BizError(&SimpleBizError{Code: Common_NotFound, Msg: msg})
//}

func InvalidParams() *SimpleBizError {
	return &SimpleBizError{Code: Common_InvalidParams}
}

func InvalidField(name string, code string, msg string) *FieldError {
	return &FieldError{SimpleBizError: &SimpleBizError{Code: code, Msg: msg}, Name: name}
}

func InvalidFieldWithMultiErrors(name string, fes *[]BizError) *FieldError {
	return &FieldError{SimpleBizError: &SimpleBizError{Errors: fes}, Name: name}
}
