//+build consul

package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"code.letsit.cn/go/common"
	"code.letsit.cn/go/common/errors"
	"code.letsit.cn/go/common/log"
	"code.letsit.cn/go/common/rpc"
	"code.letsit.cn/go/common/util"
	"code.letsit.cn/go/common/validator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

type DomainCreator func(c *gin.Context) (interface{}, error)
type IdDomainCreator func(c *gin.Context) (common.IdInf, error)
type FilterCreator func(c *gin.Context) (common.Filter, error)

type EndpointType int8

type IEndPoint interface {
	Do(c *gin.Context)
	Register(router *gin.Engine, handlers ...gin.HandlerFunc)
	SetType(typ EndpointType)
	SetRightChecker(checker func(c *gin.Context, ep *Endpoint) bool)
}

type Endpoint struct {
	Module       *common.Module
	Type         EndpointType
	HttpMethod   string
	RpcMethod    string
	Fail         func(*gin.Context, *Endpoint, error)
	Success      func(*gin.Context, *Endpoint, interface{})
	RouterPath   string
	RightKey     string
	RightChecker func(c *gin.Context, ep *Endpoint) bool
	registered   bool
	page         string
}

func (ep *Endpoint) Html() *Endpoint {
	ep.Type = END_POINT_TYPE_HTML
	ep.Success = EndpointHtmlSuccess
	ep.Fail = EndpointHtmlFail
	return ep
}

func (ep *Endpoint) Api() *Endpoint {
	ep.Type = END_POINT_TYPE_API
	ep.Success = EndpointApiSuccess
	ep.Fail = EndpointApiFail
	return ep
}

func (ep *Endpoint) SetFail(handler func(*gin.Context, *Endpoint, error)) *Endpoint {
	ep.Fail = handler
	return ep
}
func (ep *Endpoint) SetRightChecker(checker func(c *gin.Context, ep *Endpoint) bool) {
	ep.RightChecker = checker
}

func (ep *Endpoint) AlwaysPassRightCheck() *Endpoint {
	ep.RightChecker = AlwaysPassRightChecker
	return ep
}

func (ep *Endpoint) AdminRightCheck() *Endpoint {
	ep.RightChecker = EndpointAdminRightChecker
	return ep
}

func (ep *Endpoint) hasRight(c *gin.Context) bool {
	rights := c.GetStringMap(common.UserRightKey)

	if rights == nil && rights[ep.RightKey] == nil {
		return false
	}

	return true
}

func (ep *Endpoint) SetSuccess(handler func(*gin.Context, *Endpoint, interface{})) *Endpoint {
	ep.Success = handler
	return ep
}

func (ep *Endpoint) SetPath(path string) *Endpoint {
	ep.RouterPath = path
	return ep
}

func (ep *Endpoint) SetPage(page string) *Endpoint {
	ep.page = page
	return ep
}

func (ep *Endpoint) SetRpcMethod(rpcMethod string) *Endpoint {
	ep.RpcMethod = rpcMethod
	return ep
}

func (ep *Endpoint) SetType(typ EndpointType) {
	ep.Type = typ

	if typ == END_POINT_TYPE_HTML {
		if ep.Success == nil {
			ep.Success = EndpointHtmlSuccess
		}
		if ep.Fail == nil {
			ep.Fail = EndpointHtmlFail
		}
	} else {
		if ep.Success == nil {
			ep.Success = EndpointApiSuccess
		}
		if ep.Fail == nil {
			ep.Fail = EndpointApiFail
		}
	}
}

func (ep *Endpoint) path(defaultPath string) string {
	var path = ep.RouterPath
	if path == "" {
		path = defaultPath
	}
	return ep.Module.RoutePrefix + path
}

func (bep *Endpoint) validateId(c *gin.Context) (id int64, err error) {
	if err = validator.ValidateVar(validator.RULESET_ID_STRING_REQ_INT, "", c.Param("id")); err != nil {
		return
	}

	id, err = strconv.ParseInt(c.Param("id"), 10, 64)

	if err != nil {
		err = errors.InvalidParams().AddError(errors.InvalidField("id", "", "bad id format"))
	}
	return
}

func (ep *Endpoint) BindAndValidate(c *gin.Context, domain interface{}, ruleSetName string) (err error) {
	err = c.ShouldBind(domain)
	if err != nil {
		return error(&errors.SimpleBizError{Code: errors.Common_InvalidParams, Msg: err.Error()})
	}

	return validator.ValidateStruct(domain, ruleSetName)
}

func (ep *Endpoint) Bind(c *gin.Context, domain interface{}) (err error) {
	err = c.ShouldBind(domain)
	if err != nil {
		return error(&errors.SimpleBizError{Code: errors.Common_InvalidParams, Msg: err.Error()})
	}

	return nil
}

func (ep *Endpoint) register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	if ep.registered {
		return
	}
	var path = ep.Module.RoutePrefix + ep.RouterPath

	if ep.Type == END_POINT_TYPE_HTML {
		if ep.Fail == nil {
			ep.Fail = EndpointHtmlFail
		}

		if ep.Success == nil {
			ep.Success = EndpointHtmlSuccess
		}
	} else {
		if ep.Fail == nil {
			ep.Fail = EndpointApiFail
		}

		if ep.Success == nil {
			ep.Success = EndpointApiSuccess
		}
	}

	if ep.RightChecker == nil {
		ep.RightChecker = AlwaysRejectRightChecker
	}

	var httpMethod = strings.ToLower(ep.HttpMethod)
	hds := append(handlers)
	if httpMethod == "get" {
		router.GET(path, hds...)
	} else if httpMethod == "post" {
		router.POST(path, hds...)
	} else if httpMethod == "put" {
		router.PUT(path, hds...)
	} else if httpMethod == "delete" {
		router.DELETE(path, hds...)
	} else {
		panic("unsupported endpoint http method: " + httpMethod + ", " + ep.Module.Name)
	}
	ep.registered = true
}

func EndpointApiFail(c *gin.Context, md *Endpoint, err error) {
	log.Logger.Warn(md.Module.Name+" endpoint error", zap.String("error", err.Error()))

	ApiFail(c, err)
}

func EndpointApiSuccess(c *gin.Context, md *Endpoint, data interface{}) {
	c.JSON(http.StatusOK, data)
}

//TODO process error
func EndpointHtmlFail(c *gin.Context, md *Endpoint, err error) {
	log.Logger.Warn(md.Module.Name+" endpoint error", zap.String("error", err.Error()))

	HtmlFail(c, md.page, err)
}

func EndpointHtmlSuccess(c *gin.Context, md *Endpoint, data interface{}) {
	c.HTML(http.StatusOK, md.page, data)
}

var AlwaysPassRightChecker = func(c *gin.Context, ep *Endpoint) bool { return true }
var AlwaysRejectRightChecker = func(c *gin.Context, ep *Endpoint) bool { return false }

var EndpointAdminRightChecker = func(c *gin.Context, ep *Endpoint) bool {
	return AdminRightChecker(c, ep.Module.Name+":"+ep.RightKey)
}
var AdminRightChecker = func(c *gin.Context, rightKey string) bool {
	session := sessions.Default(c)
	value := session.Get("rights")
	if value == nil {
		return false
	}
	rights := value.(string)
	array := strings.Split(rights, ",")
	if len(array) < 1 {
		return false
	}
	for _, right := range array {
		if rightKey == right {
			return true
		}
	}

	return false
}

type Get struct {
	*Endpoint
	DomainCreator DomainCreator
}

func (ep *Get) Register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	if ep.DomainCreator == nil {
		panic("domain creator needed for Get endpoint")
	}

	ep.Endpoint.register(router, append(handlers, ep.Do)...)
}

func (ep *Get) Creator(domainCreator DomainCreator) *Get {
	ep.DomainCreator = domainCreator
	return ep
}

func (ep *Get) Do(c *gin.Context) {
	if !ep.RightChecker(c, ep.Endpoint) {
		ep.Fail(c, ep.Endpoint, errors.Unauthorized())
		return
	}
	var err error
	var id int64

	if id, err = ep.validateId(c); err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	//domain := rest.Domain.Domain()
	result := &common.Result{
		Error: &errors.SimpleBizError{},
	}
	result.Data, err = ep.DomainCreator(c)
	if err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}
	//log.Logger.Debug("rpc params", zap.Any("domain", domain))

	if ep.Module.RpcOn {
		if err = rpc.Call(context.Background(), ep.Module.Name, ep.RpcMethod, id, result); err != nil {
			//log.Logger.Error("failed to call rpc", zap.Any("error", err))
			ep.Fail(c, ep.Endpoint, err)
			return
		}
	} else {
		_, err := ep.Module.Db.ID(id).Get(result.Data)
		if err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
		result.Ok = true
	}

	if result.Ok {
		result.Error = nil
		ep.Success(c, ep.Endpoint, result)
		//c.JSON(http.StatusOK, result)
	} else {
		ep.Fail(c, ep.Endpoint, result.Error)
	}
}

type Create struct {
	DomainCreator DomainCreator
	*Endpoint
}

func (ep *Create) Register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	if ep.DomainCreator == nil {
		panic("domain creator needed for Create endpoint")
	}
	ep.Endpoint.register(router, append(handlers, ep.Do)...)
}

func (ep *Create) Creator(domainCreator DomainCreator) *Create {
	ep.DomainCreator = domainCreator
	return ep
}

func (ep *Create) Do(c *gin.Context) {
	log.Logger.Debug("create " + ep.Module.Name)
	if !ep.RightChecker(c, ep.Endpoint) {
		ep.Fail(c, ep.Endpoint, errors.Unauthorized())
		return
	}
	var err error

	result := &common.Result{
		Error: &errors.SimpleBizError{},
	}
	result.Data, err = ep.DomainCreator(c)
	if err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	if err = ep.BindAndValidate(c, result.Data, ""); err != nil {
		//log.Logger.Warn("failed to bind domain", zap.Any("error", err))
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	if ep.Module.RpcOn {
		if err = rpc.Call(context.Background(), ep.Module.Name, ep.RpcMethod, result.Data, result); err != nil {
			//log.Logger.Error("failed to call rpc", zap.Any("error", err))
			ep.Fail(c, ep.Endpoint, err)
			return
		}
	} else {
		_, err = ep.Module.Db.Insert(result.Data)
		if err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
		result.Ok = true
	}

	if result.Ok {
		result.Error = nil
		ep.Success(c, ep.Endpoint, result)
	} else {
		ep.Fail(c, ep.Endpoint, result.Error)
	}
}

type Update struct {
	DomainCreator IdDomainCreator
	*Endpoint
}

func (ep *Update) Register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	if ep.DomainCreator == nil {
		panic("domain creator needed for Update endpoint")
	}
	ep.Endpoint.register(router, append(handlers, ep.Do)...)
}

func (ep *Update) Creator(domainCreator IdDomainCreator) *Update {
	ep.DomainCreator = domainCreator
	return ep
}

func (ep *Update) Do(c *gin.Context) {
	log.Logger.Debug("update " + ep.Module.Name)
	if !ep.RightChecker(c, ep.Endpoint) {
		ep.Fail(c, ep.Endpoint, errors.Unauthorized())
		return
	}
	var err error

	result := &common.Result{
		Error: &errors.SimpleBizError{},
	}
	dm, err := ep.DomainCreator(c)
	if err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	if err = ep.BindAndValidate(c, dm, ""); err != nil {
		//log.Logger.Warn("failed to bind domain", zap.Any("error", err))
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	if ep.Module.RpcOn {
		if err = rpc.Call(context.Background(), ep.Module.Name, ep.RpcMethod, dm, result); err != nil {
			//log.Logger.Error("failed to call rpc", zap.Any("error", err))
			ep.Fail(c, ep.Endpoint, err)
			return
		}
	} else {
		_, err = ep.Module.Db.ID(dm.GetId()).Update(dm)
		if err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
		result.Ok = true
	}

	if result.Ok {
		result.Error = nil
		ep.Success(c, ep.Endpoint, result)
	} else {
		ep.Fail(c, ep.Endpoint, result.Error)
	}
}

type List struct {
	FilterCreator FilterCreator
	ArrayCreator  DomainCreator
	*Endpoint
}

func (ep *List) Register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	if ep.FilterCreator == nil {
		panic("domain creator needed for List endpoint")
	}
	ep.Endpoint.register(router, append(handlers, ep.Do)...)
}

func (ep *List) Creator(filterCreator FilterCreator, arrayCreator DomainCreator) *List {
	ep.FilterCreator = filterCreator
	ep.ArrayCreator = arrayCreator
	return ep
}

func (ep *List) Do(c *gin.Context) {
	log.Logger.Debug("list " + ep.Module.Name)
	if !ep.RightChecker(c, ep.Endpoint) {
		ep.Fail(c, ep.Endpoint, errors.Unauthorized())
		return
	}

	filter, err := ep.FilterCreator(c)
	if err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	if err = ep.Bind(c, filter); err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	var result = &common.FilterResult{
		Page: &common.Page{},
		Result: common.Result{
			Error: &errors.SimpleBizError{},
		},
	}
	arr, err := ep.ArrayCreator(c)
	if err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}
	result.Data = arr

	if ep.Module.RpcOn {
		if err = rpc.Call(context.Background(), ep.Module.Name, ep.RpcMethod, filter, result); err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
	} else {
		sess := ep.Module.Db.NewSession()
		filter.Apply(sess)
		count, err := sess.Table(ep.Module.TableName).FindAndCount(arr)

		if err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
		result.Page = filter.GetPage()
		result.Page.Cnt = int64(count)
		result.Ok = true
	}

	if result.Ok {
		result.Error = nil
		ep.Success(c, ep.Endpoint, result)
	} else {
		ep.Fail(c, ep.Endpoint, result.Error)
	}
}

type Delete struct {
	*Endpoint
}

func (ep *Delete) Register(router *gin.Engine, handlers ...gin.HandlerFunc) {
	ep.Endpoint.register(router, append(handlers, ep.Do)...)
}

func (ep *Delete) Do(c *gin.Context) {
	log.Logger.Debug("logically delete " + ep.Module.Name)
	if !ep.RightChecker(c, ep.Endpoint) {
		ep.Fail(c, ep.Endpoint, errors.Unauthorized())
		return
	}
	var err error
	var id int64

	if id, err = ep.validateId(c); err != nil {
		ep.Fail(c, ep.Endpoint, err)
		return
	}

	var result = &common.Result{
		Error: &errors.SimpleBizError{},
	}
	if ep.Module.RpcOn {
		if err = rpc.Call(context.Background(), ep.Module.Name, ep.RpcMethod, &id, result); err != nil {
			//log.Logger.Error("failed to call rpc", zap.Any("error", err))
			ep.Fail(c, ep.Endpoint, err)
			return
		}
	} else {
		log.Logger.Debug("delete "+ep.Module.Name+" with id ", zap.Int64("", id))
		_, err = ep.Module.Db.Exec("update "+ep.Module.TableName+" set status = 0 where id = ?", id)
		if err != nil {
			ep.Fail(c, ep.Endpoint, err)
			return
		}
		result.Ok = true
	}

	if result.Ok {
		result.Error = nil
		ep.Success(c, ep.Endpoint, result)
	} else {
		ep.Fail(c, ep.Endpoint, result.Error)
	}
}

type CrudDomainFactory interface {
	Get(c *gin.Context) (interface{}, error)
	Create(c *gin.Context) (interface{}, error)
	Update(c *gin.Context) (common.IdInf, error)
	Filter(c *gin.Context) (common.Filter, error)
	List(c *gin.Context) (interface{}, error)
}

const (
	END_POINT_TYPE_HTML EndpointType = iota
	END_POINT_TYPE_API
)

type EndpointBuilder struct {
	Module       *common.Module
	endpointType EndpointType
	endPoints    map[string]IEndPoint
}

func NewEndpointBuilder(moduleName string, tableName string, routePrefix string) *EndpointBuilder {
	return &EndpointBuilder{
		Module:    common.NewModule(moduleName, tableName, routePrefix),
		endPoints: map[string]IEndPoint{},
	}
}

func NewEndpointBuilderModule(module *common.Module) *EndpointBuilder {
	return &EndpointBuilder{
		Module:    module,
		endPoints: map[string]IEndPoint{},
	}
}

func (builder *EndpointBuilder) Crud(creator CrudDomainFactory) *EndpointBuilder {
	builder.NewGet().Creator(creator.Get)
	builder.NewCreate().Creator(creator.Create)
	builder.NewList().Creator(creator.Filter, creator.List)
	builder.NewUpdate().Creator(creator.Update)
	builder.NewDelete()
	return builder
}

func (builder *EndpointBuilder) RpcOn() *EndpointBuilder {
	builder.Module.RpcOn = true
	return builder
}

func (builder *EndpointBuilder) Rpc(on bool) *EndpointBuilder {
	builder.Module.RpcOn = on
	return builder
}

func (builder *EndpointBuilder) Api() *EndpointBuilder {
	builder.endpointType = END_POINT_TYPE_API
	builder.applyType()
	return builder
}

func (builder *EndpointBuilder) Html() *EndpointBuilder {
	builder.endpointType = END_POINT_TYPE_HTML
	builder.applyType()
	return builder
}

func (builder *EndpointBuilder) SetRightChecker(checker func(c *gin.Context, ep *Endpoint) bool) *EndpointBuilder {
	for _, iep := range builder.endPoints {
		iep.SetRightChecker(checker)
	}
	return builder
}

func (builder *EndpointBuilder) AlwaysPassRightCheck() *EndpointBuilder {
	for _, iep := range builder.endPoints {
		iep.SetRightChecker(AlwaysPassRightChecker)
	}
	return builder
}

func (builder *EndpointBuilder) AdminRightCheck() *EndpointBuilder {
	for _, iep := range builder.endPoints {
		iep.SetRightChecker(EndpointAdminRightChecker)
	}
	return builder
}

func (builder *EndpointBuilder) applyType() {
	for _, iep := range builder.endPoints {
		iep.SetType(builder.endpointType)
	}
}

func (builder *EndpointBuilder) Db(dbc *xorm.Engine) *EndpointBuilder {
	builder.Module.Db = dbc
	return builder
}

func (builder *EndpointBuilder) GetEndpoint(name string) IEndPoint {
	return builder.endPoints[name]
}

func (builder *EndpointBuilder) NewGet(endpointName ...string) *Get {
	name := "Get"
	if len(endpointName) > 0 {
		name = endpointName[0]
	}

	var ep = &Get{
		Endpoint: &Endpoint{
			Module:     builder.Module,
			HttpMethod: "Get",
			RpcMethod:  "Get",
			RouterPath: "/:id",
			RightKey:   "get",
		},
	}
	ep.page = util.LowerFirst(builder.Module.Name) + "Detail.html"
	builder.endPoints[name] = IEndPoint(ep)
	return ep
}

func (builder *EndpointBuilder) NewCreate(endpointName ...string) *Create {
	name := "Create"
	if len(endpointName) > 0 {
		name = endpointName[0]
	}

	var ep = &Create{
		Endpoint: &Endpoint{
			Module:     builder.Module,
			HttpMethod: "Post",
			RpcMethod:  "Create",
			RouterPath: "",
			RightKey:   "create",
		},
	}
	builder.endPoints[name] = IEndPoint(ep)
	return ep
}

func (builder *EndpointBuilder) NewList(endpointName ...string) *List {
	name := "List"
	if len(endpointName) > 0 {
		name = endpointName[0]
	}

	var ep = &List{
		Endpoint: &Endpoint{
			Module:     builder.Module,
			HttpMethod: "Get",
			RpcMethod:  "List",
			RouterPath: "",
			RightKey:   "list",
		},
	}
	ep.page = util.LowerFirst(builder.Module.Name) + "List.html"
	builder.endPoints[name] = IEndPoint(ep)
	return ep
}

func (builder *EndpointBuilder) NewUpdate(endpointName ...string) *Update {
	name := "Update"
	if len(endpointName) > 0 {
		name = endpointName[0]
	}

	var ep = &Update{
		Endpoint: &Endpoint{
			Module:     builder.Module,
			HttpMethod: "Put",
			RpcMethod:  "Update",
			RouterPath: "/:id",
			RightKey:   "update",
		},
	}
	builder.endPoints[name] = IEndPoint(ep)
	return ep
}

func (builder *EndpointBuilder) NewDelete(endpointName ...string) *Delete {
	name := "Delete"
	if len(endpointName) > 0 {
		name = endpointName[0]
	}

	var ep = &Delete{
		Endpoint: &Endpoint{
			Module:     builder.Module,
			HttpMethod: "Delete",
			RpcMethod:  "Delete",
			RouterPath: "/:id",
			RightKey:   "delete",
		},
	}
	builder.endPoints[name] = IEndPoint(ep)
	return ep
}

func (builder *EndpointBuilder) RegisterAll(router *gin.Engine, handlers ...gin.HandlerFunc) {
	for _, ep := range builder.endPoints {
		ep.Register(router, handlers...)
	}
}

func EnsureUid(c *gin.Context, uid *int64) error {
	*uid = c.GetInt64(common.UserIdKey)
	if *uid <= 0 {
		return errors.Unauthorized()
	}

	return nil
}
