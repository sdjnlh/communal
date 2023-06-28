package rule

import (
	be "errors"
	//"go.uber.org/zap"
	"github.com/sdjnlh/communal/errors"
	//. "github.com/sdjnlh/communal/log"
	"reflect"
	"strconv"
)

type RangeRule struct {
	BaseRule
	min float64
	max float64
}

func (rule *RangeRule) SetParams(params []string) error {
	//Logger.Debug("set parameters for validation rule ", zap.String("name", rule.BaseRule.code), zap.Any("params", params))

	if len(params) != 2 {
		return be.New("wrong parameters count set for range validation rule, expect 2 but got " + strconv.Itoa(len(params)))
	}

	var err error
	rule.min, err = strconv.ParseFloat(params[0], 64)

	if err != nil {
		return be.New("wrong parameters 'min' set for range validation rule " + params[0])
	}

	rule.max, err = strconv.ParseFloat(params[1], 64)

	if err != nil {
		return be.New("wrong parameters 'max' set for range validation rule " + params[1])
	}

	return nil
}

func (rule *RangeRule) Validate(name string, data interface{}) *errors.FieldError {
	//Logger.Debug("validate field", zap.String("name", name), zap.Any("data", data))

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &errors.FieldError{
				Name: name,
				SimpleBizError: &errors.SimpleBizError{
					Code: rule.BaseRule.code,
					Msg:  rule.BaseRule.message,
				},
			}
		} else {
			v = v.Elem()
		}
	}

	var floatVal float64

	switch v.Kind() {
	case reflect.Int:
		floatVal = float64(data.(int))
		break
	case reflect.Int8:
		floatVal = float64(data.(int8))
		break
	case reflect.Int16:
		floatVal = float64(data.(int16))
		break
	case reflect.Int32:
		floatVal = float64(data.(int32))
		break
	case reflect.Int64:
		floatVal = float64(data.(int64))
		break
	case reflect.Float32:
		floatVal = float64(data.(float32))
		break
	case reflect.Float64:
		floatVal = data.(float64)
		break
	case reflect.Ptr:
		return rule.Validate(name, v.Elem().Interface())
	default:
		panic("try to validate an unsupported data type with range validation rule " + v.Kind().String())
	}

	if floatVal >= rule.min && floatVal <= rule.max {
		return nil
	}

	return &errors.FieldError{
		Name: name,
		SimpleBizError: &errors.SimpleBizError{
			Code: rule.BaseRule.code,
			Msg:  rule.BaseRule.message,
		},
	}
}
