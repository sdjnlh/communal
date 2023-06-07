package rule

import (
	"github.com/asaskevich/govalidator"
	//"go.uber.org/zap"
	"code.letsit.cn/go/common/errors"
	be "errors"
	//. "code.letsit.cn/go/common/log"
	"strings"
)

type Rule interface {
	Validate(name string, data interface{}) *errors.FieldError
}

type ParamableRule interface {
	Rule
	SetParams([]string) error
}

type BaseRule struct {
	code    string
	message string
}

func NewRule(ruleTag string) (rule *Rule, err error) {
	var tag, key string
	var params []string
	if tag = strings.TrimSpace(ruleTag); tag == "" {
		return nil, be.New("try to create validation rule with empty tag")
	}

	bracketIndex := strings.Index(tag, "(")

	if bracketIndex == -1 {
		key = tag
	} else {
		key = tag[0:bracketIndex]
		params = strings.Split(tag[bracketIndex+1:len(tag)-1], "|")
	}
	//Logger.Debug("tag metas", zap.String("key", key), zap.Any("params", params))

	if ru, ok := fixedRules[key]; ok {
		rule = &ru
		return
	}

	if key == "range" {
		rangeRule := &RangeRule{
			BaseRule: BaseRule{
				code:    key,
				message: "该字段取值范围为" + params[0] + "-" + params[1],
			},
		}

		err = rangeRule.SetParams(params)

		if err != nil {
			return nil, err
		}

		ru := Rule(rangeRule)
		rule = &ru
		return
	}

	return nil, be.New("the validation tag is not supported: " + ruleTag)
}

//TODO retrieve and translate message based on config files
var fixedRules map[string]Rule = map[string]Rule{
	"required": &RequiredRule{
		BaseRule: BaseRule{
			code:    "empty",
			message: "该字段不能为空",
		},
	},
	"email": &StringRule{
		BaseRule: BaseRule{
			code:    "notEmail",
			message: "无效的Email地址",
		},
		validator: govalidator.IsEmail,
	},
	"numeric": &StringRule{
		BaseRule: BaseRule{
			code:    "notNumber",
			message: "无效的数字",
		},
		validator: govalidator.IsNumeric,
	},
	"int": &StringRule{
		BaseRule: BaseRule{
			code:    "notInt",
			message: "无效的整数",
		},
		validator: govalidator.IsInt,
	},
	"url": &StringRule{
		BaseRule: BaseRule{
			code:    "notUrl",
			message: "无效的URL",
		},
		validator: govalidator.IsURL,
	},
}
