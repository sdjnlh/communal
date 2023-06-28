package validator

import (
	be "errors"
	"fmt"
	"github.com/sdjnlh/communal/errors"
	"github.com/sdjnlh/communal/log"
	"github.com/sdjnlh/communal/validator/rule"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v8"
	"reflect"
	//"go.uber.org/zap"
	"strings"
	"unicode"
)

type Field struct {
	Name string
	Data interface{}
}

func NewField(name string, data interface{}) Field {
	return Field{
		Name: name,
		Data: data,
	}
}

type FieldRules struct {
	Name  string
	Rules []rule.Rule
}

var (
	RuleSetCache    map[string]map[string]FieldRules
	RuleCache       map[string]rule.Rule
	CachedValidator = CacheableValidator{}
)

type CacheableValidator struct {
}

func (dv *CacheableValidator) ValidateStruct(target interface{}, ruleSetName string) error {
	if target == nil {
		return nil
	}
	var err error
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// we only accept structs
	if val.Kind() != reflect.Struct {
		return errors.Empty()
	}

	rt := val.Type()
	rsn := formatRuleSetName(rt.PkgPath()+"."+rt.Name(), ruleSetName)

	var ruleSet = RuleSetCache[rsn]

	if ruleSet == nil {
		//TODO cache rule set even if no field rules configured
		//log.Logger.Debug("cache validation rules " + ruleSetName)
		for i := 0; i < rt.NumField(); i = i + 1 {
			fd := rt.Field(i)
			tag := fd.Tag
			//log.Logger.Debug("validate", zap.String("valid tag", tag.Get("valid")))

			if vtag := tag.Get("valid"); vtag != "" && vtag != "-" {
				//log.Logger.Debug("valid tag", zap.String("field", fd.Name), zap.String("tag", vtag))
				//fmt.Println("valid tag " + vtag)

				if err = translateAndCacheRules(rt.PkgPath()+"."+rt.Name(), rt.Field(i).Name, vtag); err != nil {
					//log.Logger.Error("translate validation rules error", zap.Any("error", err.Error()))
					panic("failed to translate validate rules " + rt.PkgPath() + "." + rt.Name())
				}
			}
		}

		ruleSet = RuleSetCache[rsn]
	} else {
		//log.Logger.Debug("use cached validation rules " + ruleSetName)
	}

	if ruleSet == nil {
		//log.Logger.Warn("nil rule set", zap.String("name", ruleSetName), zap.Any("cache", RuleSetCache))
		panic("try to validate struct without correct validating rules configured " + rt.PkgPath() + "." + rt.Name())
	} else {
		var verr = errors.InvalidParams()

		for name, fieldRules := range ruleSet {
			field := val.FieldByName(name)

			for _, ru := range fieldRules.Rules {
				if fieldErr := ru.Validate(fieldRules.Name, field.Interface()); fieldErr != nil {
					log.Logger.Info("valid field fail", zap.String("field", name), zap.String("error", fieldErr.Error()))
					verr.AddError(fieldErr)
				} else {
					log.Logger.Debug("valid tag success", zap.String("field", name))
				}
			}
		}

		if verr.HasError() {
			//log.Logger.Debug("validate error", zap.Bool("has error", verr.HasError()))
			//for _, fe := range *verr.Errors {
			//	cfe, ase := fe.(*errors.FieldError)
			//
			//	if ase != true {
			//		log.Logger.Debug("assert error", zap.Any("type", reflect.TypeOf(fe)))
			//	}
			//	log.Logger.Debug("field error", zap.String("field", cfe.Name), zap.String("error", fe.Error()))
			//}
			return verr
		}
	}

	return err
}

func ValidateStruct(target interface{}, ruleSetName string) error {
	return CachedValidator.ValidateStruct(target, ruleSetName)
}

func translateAndCacheRules(prefix string, fieldName string, rulesTag string) (err error) {
	//fmt.Println("translate ruleset field " + prefix + " " + fieldName + " " + rulesTag)
	rs := strings.TrimSpace(rulesTag)
	if rs == "" {
		return
	}

	nameRules := strings.Split(rs, "+")

	var names, rulesStr string
	var nameArr []string

	for _, nameRule := range nameRules {
		nameRuleArr := strings.Split(nameRule, "~")

		if len(nameRuleArr) > 1 {
			names = nameRuleArr[0]
			rulesStr = nameRuleArr[1]
		} else {
			rulesStr = nameRuleArr[0]
		}

		if names != "" {
			nameArr = strings.Split(names, ",")
		}

		ruleStrArr := strings.Split(rulesStr, ",")

		var rules = FieldRules{Name: fieldName, Rules: []rule.Rule{}}

		for _, ruleStr := range ruleStrArr {
			if ru, ok := RuleCache[ruleStr]; ok {
				rules.Rules = append(rules.Rules, ru)
			} else {
				if ru, err := rule.NewRule(ruleStr); err == nil {
					rules.Rules = append(rules.Rules, *ru)
				} else {
					return err
				}
			}
		}

		if len(rules.Rules) > 0 {
			if len(nameArr) > 0 {
				for _, name := range nameArr {
					cacheFieldRules(prefix+"."+name, fieldName, rules)
				}
			} else {
				cacheFieldRules(prefix+".DEFAULT", fieldName, rules)
			}
		}
	}

	return err
}

func cacheFieldRules(ruleSetKey string, fieldName string, fieldRules FieldRules) {
	//fmt.Println("cache ruleset field " + ruleSetKey + " " + fieldName)
	structRules := RuleSetCache[ruleSetKey]

	if structRules == nil {
		structRules = map[string]FieldRules{}
		RuleSetCache[ruleSetKey] = structRules
	}

	structRules[fieldName] = fieldRules
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

func (dv *CacheableValidator) ValidateMap(ruleSetId string, data map[string]interface{}) error {
	var ruleSet = RuleSetCache[ruleSetId]

	if ruleSet == nil {
		panic("validation rule set " + ruleSetId + " not registered")
	}

	var verr = errors.InvalidParams()

	for name, fieldRules := range ruleSet {
		for _, ru := range fieldRules.Rules {
			if fieldErr := ru.Validate(fieldRules.Name, data[name]); fieldErr != nil {
				verr.AddError(fieldErr)
			}
		}
	}

	if verr.HasError() {
		return verr
	}

	return nil
}

func ValidateMap(ruleSetId string, data map[string]interface{}) error {
	return CachedValidator.ValidateMap(ruleSetId, data)
}

func (dv *CacheableValidator) ValidateVar(ruleSetNamespace string, ruleSetName string, data interface{}) error {
	key := formatRuleSetName(ruleSetNamespace, ruleSetName)

	var ruleSet = RuleSetCache[key]

	if ruleSet == nil {
		panic("validation rule set " + key + " not registered")
	}

	if len(ruleSet) != 1 {
		panic("trying to validate single variable with multiple fields rule set " + key)
	}

	var verr = errors.InvalidParams()

	for _, fieldRules := range ruleSet {
		for _, ru := range fieldRules.Rules {
			if fieldErr := ru.Validate(fieldRules.Name, data); fieldErr != nil {
				verr.AddError(fieldErr)
			}
		}
	}

	if verr.HasError() {
		return verr
	}

	return nil
}

func ValidateVar(ruleSetNamespace string, ruleSetName string, data interface{}) error {
	return CachedValidator.ValidateVar(ruleSetNamespace, ruleSetName, data)
}

func (dv *CacheableValidator) RegisterRuleSet(ruleSetNamespace string, fieldTags map[string]string) error {
	//fmt.Println("register rule set " + ruleSetNamespace)
	//this is not rule set id, just namespace
	//if ruleSet := RuleSetCache[ruleSetNamespace]; ruleSet != nil {
	//	return be.New("rule set with the same name " + ruleSetNamespace + " has been registered")
	//}

	for fieldName, tag := range fieldTags {
		if err := translateAndCacheRules(ruleSetNamespace, fieldName, tag); err != nil {
			//log.Logger.Error("translate validation rules error", zap.Any("error", err.Error()))
			fmt.Println("failed to translate validate rules " + ruleSetNamespace + ", " + err.Error())
			return be.New("failed to translate validate rules " + ruleSetNamespace + ", " + err.Error())
		}
	}

	return nil
}

func RegisterRuleSet(ruleSetNamespace string, fieldTags map[string]string) error {
	return CachedValidator.RegisterRuleSet(ruleSetNamespace, fieldTags)
}

func (dv *CacheableValidator) RegisterValidation(ruleSetNamespace string, f validator.Func) error {
	return nil
}

func formatRuleSetName(prefix string, name string) string {
	if strings.TrimSpace(name) == "" {
		return prefix + ".DEFAULT"
	} else {
		return prefix + "." + name
	}
}

const (
	RULESET_ID_STRING_REQ_INT = "letsit.cn/ruleset/id.string_req_int"
	RULESET_ID_SKIP           = "skip"
)

func registerCommonRulesets() {
	RegisterRuleSet(RULESET_ID_STRING_REQ_INT, map[string]string{"id": "required,int"})
}

func init() {
	fmt.Println("init validator")
	RuleSetCache = make(map[string]map[string]FieldRules)
	registerCommonRulesets()
}
