package web

import (
	"code.letsit.cn/go/common"
	"code.letsit.cn/go/common/errors"
	"code.letsit.cn/go/common/validator"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type HtmlHandler struct {
	*Handler
}

var DefaultHtmlHandler = &HtmlHandler{}

//func (handler *HtmlHandler) Error(c *gin.Context, code int, data interface{}, redirect bool) {
//	c.AbortWithStatusJSON(http.StatusInternalServerError, err)
//}

func (handler *HtmlHandler) BadRequest(c *gin.Context, page string, data interface{}) {
	if page == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/400.html")
		return
	}

	if be, ok := data.(*errors.SimpleBizError); ok {
		c.HTML(http.StatusBadRequest, page, be)
	} else {
		c.HTML(http.StatusBadRequest, page, gin.H{"ok": false, "data": data})
	}
}

func (handler *HtmlHandler) Unauthorized(c *gin.Context, page string, data interface{}) {
	if page == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login.html")
		return
	}

	if be, ok := data.(*errors.SimpleBizError); ok {
		c.HTML(http.StatusUnauthorized, page, be)
	} else {
		c.HTML(http.StatusUnauthorized, page, gin.H{"ok": false, "data": data})
	}
}

func (handler *HtmlHandler) NotFound(c *gin.Context, page string, data interface{}) {
	if page == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/404.html")
		return
	}

	if be, ok := data.(*errors.SimpleBizError); ok {
		c.HTML(http.StatusNotFound, page, be)
	} else {
		c.HTML(http.StatusNotFound, page, gin.H{"ok": false, "data": data})
	}
}

func (handler *HtmlHandler) InternalError(c *gin.Context, page string, data interface{}) {
	if page == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/500.html")
		return
	}

	if be, ok := data.(*errors.SimpleBizError); ok {
		c.HTML(http.StatusInternalServerError, page, be)
	} else {
		c.HTML(http.StatusInternalServerError, page, gin.H{"ok": false, "data": data})
	}
}
func HtmlFail(c *gin.Context, page string, err error) {
	writeResult(c, nil, err, page)
}

func writeResult(c *gin.Context, result common.IResult, err error, pages ...string) {
	if result != nil && result.IsOk() {
		c.HTML(http.StatusOK, pages[0], result)
		return
	}

	errPage := ""
	if len(pages) == 2 {
		errPage = pages[1]
	}

	code := http.StatusInternalServerError

	if err == nil {
		if result != nil && result.Err() != nil {
			switch result.Err().GetCode() {
			case errors.Common_InvalidParams:
				code = http.StatusBadRequest
			case errors.Common_NotFound:
				code = http.StatusNotFound
			case errors.Common_Forbidden:
				code = http.StatusForbidden
			default:
				break
			}
		}
	}

	if errPage == "" {
		if code == http.StatusBadRequest {
			errPage = "/400.html"
		} else if code == http.StatusNotFound {
			errPage = "/404.html"
		} else {
			errPage = "/500.html"
		}
		c.Redirect(http.StatusTemporaryRedirect, errPage)
		return
	} else {
		c.HTML(code, errPage, result)
		return
	}
}

func (handler *HtmlHandler) Result(c *gin.Context, page string, result common.IResult) {
	writeResult(c, result, nil, page)
}

func (handler *HtmlHandler) ResultWithError(c *gin.Context, result common.IResult, err error, pages ...string) {
	writeResult(c, result, err, pages...)
}

func (handler *HtmlHandler) ValidateInt64Id(c *gin.Context) (id int64, err error) {
	if err = validator.ValidateVar(validator.RULESET_ID_STRING_REQ_INT, "", c.Param("id")); err != nil {
		//c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	id, err = strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		//c.AbortWithStatusJSON(http.StatusBadRequest, errors.InvalidParams().AddError(errors.InvalidField("id", "", "bad id format")))
	}
	return
}

func (handler *HtmlHandler) MustUid(c *gin.Context) (uid int64, ok bool) {
	uid = c.GetInt64(common.UserIdKey)
	if uid <= 0 {
		handler.Unauthorized(c, "", nil)
		return uid, false
	}

	return uid, true
}
