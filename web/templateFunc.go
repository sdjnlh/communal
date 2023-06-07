package web

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
	//ginJson "github.com/gin-gonic/gin/json"
	"encoding/json"
)

func Timing(t time.Time, format string) string {
	if t.IsZero() {
		return ""
	}

	if format == "" {
		return t.Format("2006-01-02 15:04:05")
	} else {
		return t.Format(format)
	}
}

func Int64Divide(value int64, factor int) string {
	return strconv.FormatInt(value/int64(factor), 10)
}

func Float64Divide(value float64, factor int) string {
	return strconv.FormatInt(int64(value/float64(factor)), 10)
}

func unescaped(x string) interface{} { return template.HTML(x) }

/**
*	not nil
 */
func notNil(x interface{}) bool { return x != nil }

func marshal(value interface{}) template.JS {
	bts, _ := json.Marshal(value)

	return template.JS(bts)
}

func MapValue(m map[string]interface{}, key interface{}) interface{} {
	if m == nil {
		return nil
	}

	var path string
	path, ok := key.(string)
	if !ok || path == "" {
		return ""
	}

	var tm = m

	keys := strings.Split(path, ".")
	for index, key := range keys {
		if tm == nil {
			return nil
		}

		if index == len(keys)-1 {
			return tm[key]
		}

		if tm[key] == nil {
			return nil
		}
		tm = tm[key].(map[string]interface{})
	}
	return nil
}

func StringEqual(value interface{}, str string) bool {
	if v, ok := value.(string); ok {
		return v == str
	}
	return false
}

var CommonFuncMap template.FuncMap

func MergeFuncMap(fm template.FuncMap, fms ...template.FuncMap) {
	if fm == nil {
		fmt.Errorf("fail to merge FuncMap, nil target")
		return
	}

	if fms != nil {
		for _, nfm := range fms {
			for k, v := range nfm {
				fm[k] = v
			}
		}
	}
}

func init() {
	CommonFuncMap = template.FuncMap{}
	CommonFuncMap["unescaped"] = unescaped
	CommonFuncMap["timing"] = Timing
	CommonFuncMap["int64Divide"] = Int64Divide
	CommonFuncMap["float64Divide"] = Float64Divide
	CommonFuncMap["json"] = marshal
	CommonFuncMap["nn"] = notNil
	CommonFuncMap["seq"] = StringEqual
	CommonFuncMap["mpv"] = MapValue
}
