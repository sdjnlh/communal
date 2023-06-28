package web

import (
	"github.com/gin-gonic/gin"
	"github.com/sdjnlh/communal"
	"github.com/sdjnlh/communal/errors"
	"github.com/sdjnlh/communal/log"
	"github.com/sdjnlh/communal/validator"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type RestHandler struct {
	*Handler
}

var DefaultRestHandler = &RestHandler{}

func (handler *RestHandler) Success(c *gin.Context) {
	handler.SuccessWithData(c, nil)
}

func (handler *RestHandler) SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func (handler *RestHandler) SuccessWithPair(c *gin.Context, key string, value interface{}) {
	c.JSON(http.StatusOK, map[string]interface{}{key: value})
}

func (handler *RestHandler) Fail(c *gin.Context) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (handler *RestHandler) FailWithCode(c *gin.Context, code string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, errors.SimpleBizError{Code: code})
}

func (handler *RestHandler) FailWithMessage(c *gin.Context, code string, message string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, errors.SimpleBizError{Code: code, Msg: message})
}

func (handler *RestHandler) FailWithError(c *gin.Context, err error) {
	log.Logger.Warn("api fail", zap.Error(err))
	c.AbortWithStatus(http.StatusInternalServerError)
}

func (handler *RestHandler) FailWithBizError(c *gin.Context, err errors.BizError) {
	log.Logger.Warn("api fail", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusInternalServerError, err)
}

func (handler *RestHandler) RPCError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, errors.RPCFailed())
}

func (handler *RestHandler) BadRequest(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
}

func (handler *RestHandler) BadRequestWithError(c *gin.Context, err error) {
	log.Logger.Warn("api fail", zap.Error(err))
	c.AbortWithStatusJSON(http.StatusBadRequest, err)
}

func (handler *RestHandler) Unauthorized(c *gin.Context) {
	c.AbortWithStatus(http.StatusUnauthorized)
}

func (handler *RestHandler) Forbidden(c *gin.Context) {
	c.AbortWithStatus(http.StatusForbidden)
}

func (handler *RestHandler) NotFound(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotFound)
}

func (handler *RestHandler) ResultWithError(c *gin.Context, result communal.IResult, err error) {
	if err != nil {
		log.Logger.Warn("api fail", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
	} else {
		handler.Result(c, result)
	}
}

func (handler *RestHandler) Result(c *gin.Context, result communal.IResult) {
	if result == nil {
		c.AbortWithStatus(http.StatusOK)
		return
	}

	if result.IsOk() {
		result.SetError(nil)
		c.JSON(http.StatusOK, result)
		return
	}

	if result.Err() != nil {
		switch result.Err().GetCode() {
		case errors.Common_InvalidParams:
			c.AbortWithStatusJSON(http.StatusBadRequest, result.Err())
			break
		case errors.Common_NotFound:
			c.AbortWithStatusJSON(http.StatusNotFound, result.Err())
			break
		case errors.Common_Unauthorized:
			c.AbortWithStatusJSON(http.StatusUnauthorized, result.Err())
			break
		case errors.Common_Forbidden:
			c.AbortWithStatusJSON(http.StatusForbidden, result.Err())
			break
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, result.Err())
			break
		}

		return
	}

	c.AbortWithStatus(http.StatusInternalServerError)
}

func ApiFail(c *gin.Context, err error) {
	if be, ok := err.(errors.BizError); ok {
		if be.GetCode() == errors.Common_InvalidParams {
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, &communal.Result{Ok: false, Error: errors.ServerErrorWithMsg(err.Error())})
}

func (handler *RestHandler) ValidateInt64Id(c *gin.Context) (id int64, err error) {
	if err = validator.ValidateVar(validator.RULESET_ID_STRING_REQ_INT, "", c.Param("id")); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	id, err = strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errors.InvalidParams().AddError(errors.InvalidField("id", "", "bad id format")))
	}
	return
}
