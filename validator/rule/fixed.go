package rule

import (
	"github.com/asaskevich/govalidator"
	"github.com/sdjnlh/communal/errors"
	//"go.uber.org/zap"
	"github.com/sdjnlh/communal/util"
	//. "github.com/sdjnlh/communal/log"
)

type StringRule struct {
	BaseRule
	validator govalidator.Validator
}

func (rule *StringRule) Validate(key string, data interface{}) *errors.FieldError {
	//Logger.Debug("validate field with rule " + rule.code, zap.String("name", key), zap.Any("data", data))

	str, err := util.EnsureString(data)

	if err != nil {
		//Logger.Error("validate filed error", zap.String("error", err.Error()))
		panic("regexp validation rule configured on a field not string type " + key)
	}

	if ok := rule.validator(str); !ok {
		return &errors.FieldError{
			Name: key,
			SimpleBizError: &errors.SimpleBizError{
				Code: rule.BaseRule.code,
				Msg:  rule.BaseRule.message,
			},
		}
	}

	return nil
}

type RequiredRule struct {
	BaseRule
}

func (rule *RequiredRule) Validate(name string, data interface{}) *errors.FieldError {
	//Logger.Debug("validate field with rule required", zap.String("name", name), zap.Any("data", data))

	value, isNil := util.Indirect(data)
	if isNil || util.IsEmpty(value) {
		return &errors.FieldError{
			Name: name,
			SimpleBizError: &errors.SimpleBizError{
				Code: rule.BaseRule.code,
				Msg:  rule.BaseRule.message,
			},
		}
	}
	return nil
}
