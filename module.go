package common

import (
	"context"
	"reflect"
	"time"

	"code.letsit.cn/go/common/errors"
	"code.letsit.cn/go/common/log"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

type DBListener interface {
	SetDB(db *xorm.Engine)
	GetDbName() string
	SetDbName(string)
	DbEnabled() bool
}

type IModule interface {
	DBListener
	GetName() string
	GetTableName() string
}

type Module struct {
	Name        string
	TableName   string
	RoutePrefix string

	RpcOn  bool
	DbOn   bool
	DbName string
	Db     *xorm.Engine
}

func NewModule(name string, tableName string, routePrefix string) *Module {
	return &Module{
		Name:        name,
		TableName:   tableName,
		RoutePrefix: routePrefix,
		DbOn:        true,
	}
}

func (module *Module) SetDB(db *xorm.Engine) {
	module.Db = db
}

func (module *Module) GetDbName() string {
	return module.DbName
}

func (module *Module) SetDbName(dbName string) {
	module.DbName = dbName
}

func (module *Module) DbEnabled() bool {
	return module.DbOn
}

func (module *Module) GetName() string {
	return module.Name
}

func (module *Module) GetTableName() string {
	return module.TableName
}

func (module *Module) EnableDb(dbName string) IModule {
	module.DbName = dbName
	module.DbOn = true
	return module
}

func (module *Module) DisableDb(dbName string) IModule {
	module.DbOn = false
	return module
}

func (module *Module) Get(ctx context.Context, i interface{}, receiver *Result, funcs ...func(ss *xorm.Session)) (err error) {
	if i == nil {
		receiver.Failure(errors.InvalidParams())
		return
	}

	var id int64
	if reflect.TypeOf(i).Kind().String() == "ptr" {
		id = reflect.Indirect(reflect.ValueOf(i)).Int()
	} else {
		id = i.(int64)
	}
	ss := module.Db.NewSession()
	ss.Table(module.GetTableName())
	if len(funcs) > 0 {
		funcs[0](ss)
	}
	if _, err = ss.ID(id).Get(receiver.Data); err != nil {
		log.Logger.Error("", zap.Error(err))
		return err
	}
	receiver.Success()
	return
}

func (module *Module) Create(ctx context.Context, domain interface{}, receiver *Result) (err error) {
	_, err = module.Db.Insert(domain)
	if err == nil {
		receiver.Success(domain)
	}

	return err
}

func (module *Module) List(ctx context.Context, filter Filter, result *FilterResult) (err error) {

	log.Logger.Debug("filter list", zap.Any("filter", filter))

	session := module.Db.Table(module.TableName).Desc("id")
	filter.Apply(session)
	count, err := session.FindAndCount(result.Data)
	if err != nil {
		return err
	}

	result.Ok = true
	result.Page = filter.GetPage()
	result.Page.Cnt = int64(count)

	return
}

type SqlSession struct {
	xorm.Session
	alias   string
	result  *FilterResult
	filter  IFilter
	reveal  bool
	showAll bool
}

func (s *SqlSession) notBeDeleted() *SqlSession {
	NotBeDeleted := "dtd=false"
	if len(s.alias) > 0 {
		NotBeDeleted = s.alias + "." + NotBeDeleted
	}
	s.Where(NotBeDeleted)
	return s
}

func (s *SqlSession) Reveal() *SqlSession {
	s.reveal = true
	return s
}

func (s *SqlSession) All() *SqlSession {
	s.showAll = true
	return s
}

func (session *SqlSession) Alias(alias string) *SqlSession {
	session.Session.Alias(alias)
	session.alias = alias
	return session
}

func (s *SqlSession) Do(condiBean ...interface{}) error {
	defer s.Close()
	if !s.reveal {
		s.notBeDeleted()
	}

	if s.showAll { // 查询所有的
		if err := s.Find(s.result.Data); err != nil {
			log.Logger.Error("", zap.Error(err))
			return err
		}
		s.result.Success()
		return nil

	}

	s.result.Page = s.filter.GetPage()
	if len(condiBean) > 0 {
		count, err := s.Limit(s.filter.GetPage().Limit(), s.filter.GetPage().Skip()).FindAndCount(s.result.Data, condiBean[0])
		if err != nil {
			return err
		}
		s.result.Page.Cnt = count
	} else {
		count, err := s.Limit(s.filter.GetPage().Limit(), s.filter.GetPage().Skip()).FindAndCount(s.result.Data)
		if err != nil {
			return err
		}
		s.result.Page.Cnt = int64(count)
	}

	s.result.Success()
	return nil
}

// Page结构体具体实现
type IFilter interface {
	GetPage() *Page
}

func (module *Module) ListCondition(ctx context.Context, filter IFilter, result *FilterResult, funcs ...func(ss *SqlSession)) *SqlSession {
	sqlSession := &SqlSession{}
	sqlSession.Session = *(module.Db.NewSession())
	sqlSession.result = result
	sqlSession.filter = filter
	if len(funcs) > 0 {
		funcs[0](sqlSession)
	}

	return sqlSession
}

func (module *Module) Update(ctx context.Context, idm IdInf, result *Result) (err error) {
	log.Logger.Debug("update ", zap.Any(module.Name, idm))
	if idm.GetId() <= 0 {
		result.Failure(errors.InvalidParams())
		return errors.InvalidParams()
	}
	if _, err = module.Db.ID(idm.GetId()).Update(idm); err != nil {
		log.Logger.Error("fail to update item", zap.Error(err))
		return err
	}
	result.Success()
	return nil
}

// Deprecated: Use Dtd instead.
func (module *Module) Delete(ctx context.Context, id *int64, result *Result) (err error) {
	return module.delete(ctx, *id, result, "status", 0)
}

func (module *Module) Dtd(ctx context.Context, id int64, result *Result) (err error) {
	return module.delete(ctx, id, result, "dtd", true)
}

func (module *Module) delete(ctx context.Context, id int64, result *Result, key string, value interface{}) (err error) {

	if id <= 0 {
		result.Failure(errors.InvalidParams())
		return
	}
	if _, err = module.Db.Exec("update   `"+module.TableName+"`  set "+key+"=? ,lut=? where id = ?", value, time.Now(), id); err != nil {
		log.Logger.Error("", zap.Error(err))
		result.Failure(errors.InvalidParams())
		return
	}
	result.Success()
	return nil
}
