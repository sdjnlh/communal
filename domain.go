package communal

import (
	"bytes"
	"encoding/json"
	be "errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/builder"

	"github.com/gin-gonic/gin"
	"github.com/sdjnlh/communal/db"
	"github.com/sdjnlh/communal/errors"
	"github.com/sdjnlh/communal/id"
	"xorm.io/xorm"
)

const (
	UserKey          = "user"
	UserIdKey        = "uid"
	UserFirstNameKey = "firstName"
	UserLastNameKey  = "lastName"
	UserEmailKey     = "email"
	UserNicknameKey  = "nickname"
	UserRoleKey      = "role"
	UserRightKey     = "rights"
	UserTypeKey      = "tp"
	UserOrgIdKey     = "orgId"
	UserGroupKey     = "group"
)

type IdInf interface {
	SetId(int64)
	GetId() int64
}

type ID struct {
	Id int64 `xorm:"pk BIGINT(20)" json:"id,string" form:"id"`
}

func (idb *ID) SetId(id int64) {
	idb.Id = id
}

func (idb *ID) GetId() int64 {
	return idb.Id
}

type Base struct {
	ID     `xorm:"extends"`
	Crt    time.Time `xorm:"default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"crt"`
	Lut    time.Time `xorm:"default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"lut"`
	Status int16     `xorm:"default 1 TINYINT(2)" json:"status" form:"status"`
}

func (base *Base) InitBaseFields() {
	id, _ := id.Next()
	//base.Idb = &IdBean{Id: id}
	base.Id = id
	now := time.Now()
	base.Crt = now
	base.Lut = now
	base.Status = db.STATUS_COMMON_OK
}

func (base *Base) InitTimeAndStatus() {
	now := time.Now()
	base.Crt = now
	base.Lut = now
	base.Status = db.STATUS_COMMON_OK
}

type DBase struct {
	ID  `xorm:"extends"`
	Crt time.Time `xorm:"default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"crt"`
	Lut time.Time `xorm:"default 'CURRENT_TIMESTAMP' TIMESTAMP" json:"lut"`
	Dtd bool      `json:"-"`
}

func (base *DBase) InitBaseFields() {
	base.Id, _ = id.Next()
	now := time.Now()
	base.Crt = now
	base.Lut = now
}

func (base *DBase) InitTimeAndStatus() {
	now := time.Now()
	base.Crt = now
	base.Lut = now
}

var DeleteDomain = map[string]interface{}{"status": db.STATUS_COMMON_DELETED}

type IResult interface {
	IsOk() bool
	Err() errors.BizError
	SetError(err errors.BizError)
	Set(key string, value interface{})
}

type Result struct {
	Ok    bool                   `json:"ok"`
	Error errors.BizError        `json:"err,omitempty"`
	Data  interface{}            `json:"data,omitempty"`
	User  interface{}            `json:"user,omitempty"`
	Extra map[string]interface{} `json:"extra,omitempty"`
}

func (r *Result) IsOk() bool {
	return r.Ok
}

func (r *Result) FillUser(c *gin.Context) {
	r.User, _ = c.Get(UserKey)
}

func (r *Result) Set(key string, value interface{}) {
	if r.Extra == nil {
		r.Extra = map[string]interface{}{}
	}
	r.Extra[key] = value
}

func (r *Result) Err() errors.BizError {
	return r.Error
}

func (r *Result) SetError(err errors.BizError) {
	r.Error = err
}

func (r *Result) Failure(errs ...errors.BizError) *Result {
	r.Ok = false
	if len(errs) > 0 {
		r.Error = errs[0]
	}
	return r
}

func (r *Result) FailureWithData(data interface{}, err errors.BizError) *Result {
	r.Ok = false
	r.Error = err
	r.Data = data

	return r
}

func (r *Result) Success(ds ...interface{}) *Result {
	r.Ok = true
	if len(ds) > 0 {
		r.Data = ds[0]
	}
	return r
}

func NewResult(data interface{}) *Result {
	return &Result{Error: &errors.SimpleBizError{}, Data: data}
}

type Page struct {
	P   int    `json:"p" form:"p"`
	Ps  int    `json:"ps" form:"ps"`
	Cnt int64  `json:"cnt" form:"cnt"`
	K   string `json:"k" form:"k"`
	Pc  int    `json:"pc" form:"pc"`
	Od  string `json:"od,omitempty" form:"od"`
}

func (page *Page) GetPage() *Page {
	return page
}

func (page *Page) GetPager(count int64) *Page {
	page.Cnt = count
	if page.P < 1 {
		page.P = 1
	}
	if page.Ps < 1 {
		page.Ps = db.DEFAULT_PAGE_SIZE
	}
	page.Pc = int(page.Cnt)/page.Ps + 1
	return page
}

func (page *Page) Skip() int {
	if page.Ps > 0 {
		return (page.P - 1) * page.Ps
	}

	return (page.P - 1) * db.DEFAULT_PAGE_SIZE
}

func (page *Page) Limit() int {
	if page.Ps > 0 {
		return page.Ps
	}

	return db.DEFAULT_PAGE_SIZE
}

type FilterResult struct {
	Result
	Page *Page `json:"page,omitempty"`
}

func NewFilterResult(data interface{}) *FilterResult {
	return &FilterResult{
		Result: Result{Error: &errors.SimpleBizError{}, Data: data},
	}
}

type PagedKeywordFilter struct {
	ColumnName string
	Keyword    string `json:"k" form:"k"`
	Page
}

func (filter *PagedKeywordFilter) Apply(session *xorm.Session) {
	session.Where(builder.Eq{"status": db.STATUS_COMMON_OK})
	if filter.Keyword != "" {
		session.And("name like ?", "%"+filter.Keyword+"%")
	}
	session.Limit(filter.Limit(), filter.Skip())
}

func NewPagedKeywordFilter(name string) *PagedKeywordFilter {
	return &PagedKeywordFilter{
		ColumnName: name,
	}
}

type KeywordFilter struct {
	ColumnName string
	Keyword    string `json:"k" form:"k"`
	Page
}

func (filter *KeywordFilter) Apply(session *xorm.Session) {
	session.Where(builder.Eq{"status": db.STATUS_COMMON_OK})
	if filter.Keyword != "" {
		session.And("name like ?", "%"+filter.Keyword+"%")
	}
	session.Limit(filter.Limit(), filter.Skip())
}

func NewKeywordFilter(name string) *KeywordFilter {
	return &KeywordFilter{
		ColumnName: name,
	}
}

type Labelable interface {
	Label() string
	SetLabel(string)
}

type Label struct {
	text string
}

func (label *Label) Label() string {
	return label.text
}

func (label *Label) SetLabel(str string) {
	label.text = str
}

type Filter interface {
	Apply(session *xorm.Session)
	GetPage() *Page
}

type Domain interface {
	Name() string
	Domain(init bool) interface{}
	Filter() Filter
	Array() interface{}
}

type SearchForm struct {
	Keyword string `json:"keyword" form:"keyword"`
	Page
}

type DomainContentForm struct {
	Id     int64 `form:"id" valid:"required"`
	Flag   int   `form:"flag"`
	UserId int64
	Page
}

type Cover struct {
	Id   int64  `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Path string `json:"fid,omitempty"`
	Url  string `json:"url,omitempty"`
}

func (c *Cover) FromDB(bytes []byte) error {
	return json.Unmarshal(bytes, c)
}

func (c *Cover) ToDB() (bytes []byte, err error) {
	bytes, err = json.Marshal(c)
	return
}

type JsonMap map[string]interface{}

func (c *JsonMap) FromDB(bytes []byte) error {
	return json.Unmarshal(bytes, c)
}

func (c *JsonMap) ToDB() (bytes []byte, err error) {
	if c == nil {
		return []byte("{}"), nil
	}
	bytes, err = json.Marshal(c)
	return
}

func (c *JsonMap) String() string {
	if c == nil {
		return ""
	}
	bts, _ := json.Marshal(c)
	return string(bts)
}

type StringArray []string

func (s *StringArray) FromDB(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}

	str := string(bts)
	if strings.HasPrefix(str, "{") {
		str = str[1:len(str)]
	}

	if strings.HasSuffix(str, "}") {
		str = str[0 : len(str)-1]
	}

	if str == "" {
		*s = []string{}
		return nil
	}
	arr := strings.Split(str, ",")

	for _, item := range arr {
		*s = append(*s, strings.TrimSuffix(strings.TrimPrefix(item, "\""), "\""))
	}
	return nil
}

func (s *StringArray) ToDB() ([]byte, error) {
	if s == nil {
		return nil, nil
	}

	var buffer bytes.Buffer
	buffer.WriteByte('{')
	if s == nil {
		buffer.WriteByte('}')
		return buffer.Bytes(), nil
	}
	for index, str := range *s {
		if index > 0 {
			buffer.WriteByte(',')
		}
		//buffer.WriteByte('"')
		buffer.WriteString(str)
		//buffer.WriteByte('"')
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func (s *StringArray) ToIn() []byte {
	return serializeStringArray(*s, "(", ")")
}

type Int64Array []int64

func (s *Int64Array) Contains(val int64) bool {
	if s == nil {
		return false
	}

	for _, v := range *s {
		if v == val {
			return true
		}
	}

	return false
}

func (s *Int64Array) Merge(ia Int64Array) {
	if s == nil {
		*s = ia
	}

	for _, v := range ia {
		if !s.Contains(v) {
			*s = append(*s, v)
		}
	}
}

func (s *Int64Array) FromDB(bts []byte) error {
	if len(bts) == 0 {
		return nil
	}

	str := string(bts)
	if strings.HasPrefix(str, "{") {
		str = "[" + str[1:len(str)]
	}

	if strings.HasSuffix(str, "}") {
		str = str[0:len(str)-1] + "]"
	}

	var ia = &[]int64{}

	err := json.Unmarshal([]byte(str), ia)
	if err != nil {
		return err
	}

	*s = Int64Array(*ia)
	return nil
}

func (s *Int64Array) ToDB() ([]byte, error) {
	return serializeBigIntArray(*s, "{", "}"), nil
}

func (s *Int64Array) ToIn() string {
	return string(serializeBigIntArray(*s, "(", ")"))
}

func (arr Int64Array) MarshalJSON() ([]byte, error) {
	return serializeBigIntArrayAsString(arr, "[", "]"), nil
}

func (arr *Int64Array) UnmarshalJSON(b []byte) error {
	var strarr []string
	var intarr []int64

	err := json.Unmarshal(b, &strarr)
	if err != nil {
		return err
	}

	for _, s := range strarr {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}

		intarr = append(intarr, i)
	}

	*arr = intarr
	return nil
}

const DATE_LAYOUT = "2006-01-02"

// Date of format yyyy-MM-dd
type UTCDate time.Time

//MarshalJSON encode to json value
func (date *UTCDate) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(*date).UTC().Format(DATE_LAYOUT) + "\""), nil
}

//UnmarshalJSON parse from json value
func (date *UTCDate) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	if str == "" {
		return nil
	}
	if len(b) != 12 {
		return be.New("invalid date value " + str + ", should be of format yyyy-MM-dd")
	}
	t, err := time.ParseInLocation(DATE_LAYOUT, string(b[1:11]), time.UTC)
	if err != nil {
		return err
	}
	// t = t.UTC()
	*date = UTCDate(t)
	return nil
}

//FromDB decode from database readout
func (date *UTCDate) FromDB(bts []byte) error {
	// fmt.Println("parse date from db", string(bts))
	var t time.Time
	var err error
	if len(bts) >= 12 {
		t, err = time.ParseInLocation(DATE_LAYOUT, string(bts[0:10]), time.UTC)
	}

	if err != nil {
		return err
	}
	// t = t.UTC()
	*date = UTCDate(t)
	return nil
}

//ToDB encode for writing to database
func (date *UTCDate) ToDB() ([]byte, error) {
	return date.MarshalJSON()
}

//UnixTime numeric format of time, ignore timezone
type UnixTime time.Time

//MarshalJSON encode to json value
func (ut *UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(time.Time(*ut).Unix(), 10)), nil
}

//UnmarshalJSON parse from json value
func (ut *UnixTime) UnmarshalJSON(b []byte) error {
	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}
	*ut = UnixTime(time.Unix(i, 0))
	return nil
}

func serializeBigIntArray(s []int64, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(strconv.FormatInt(val, 10))
	}

	buffer.WriteString(suffix)

	return buffer.Bytes()
}

func serializeBigIntArrayAsString(s []int64, prefix string, suffix string) []byte {
	var buffer bytes.Buffer

	buffer.WriteString(prefix)

	for idx, val := range s {
		if idx > 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteByte('"')
		buffer.WriteString(strconv.FormatInt(val, 10))
		buffer.WriteByte('"')
	}
	buffer.WriteString(suffix)
	return buffer.Bytes()
}

func serializeStringArray(s []string, prefix string, suffix string) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(prefix)

	for index, str := range s {
		if index > 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteByte('"')
		buffer.WriteString(str)
		buffer.WriteByte('"')
	}
	buffer.WriteString(suffix)
	return buffer.Bytes()
}

type Context map[string]interface{}

func (ctx *Context) Get(key string) interface{} {
	return (*ctx)[key]
}

func (ctx *Context) MustGet(key string) interface{} {
	v := (*ctx)[key]

	if v == nil {
		panic("key " + key + " not present in context")
	}
	return v
}

func (ctx *Context) Set(key string, value interface{}) {
	(*ctx)[key] = value
}

type TreeNode interface {
	GetId() int64
	GetParentId() int64
	AddChild(node TreeNode)
}

type TreeBuilder struct {
	cache map[int64]interface{}
}

const RootsNodeKey = 0

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{
		cache: map[int64]interface{}{RootsNodeKey: []TreeNode{}},
	}
}

func (builder *TreeBuilder) Node(node TreeNode) {
	n, ok := node.(TreeNode)
	if !ok {
		return
	}

	if builder.cache[n.GetId()] == nil {
		childrenKey := n.GetId() * -1
		if children, ok := builder.cache[childrenKey].([]TreeNode); ok {
			for _, child := range children {
				n.AddChild(child)
			}
		}
		builder.cache[n.GetId()] = n
	}

	if n.GetParentId() <= 0 {
		roots, ok := builder.cache[RootsNodeKey].([]TreeNode)
		if ok {
			roots = append(roots, n)
		} else {
			roots = []TreeNode{n}
		}
		builder.cache[RootsNodeKey] = roots
		return
	}

	if pn, ok := builder.cache[n.GetParentId()].(TreeNode); ok {
		pn.AddChild(n)
	} else {
		childrenKey := n.GetParentId() * -1
		if builder.cache[childrenKey] == nil {
			builder.cache[childrenKey] = []TreeNode{n}
		} else {
			builder.cache[childrenKey] = append(builder.cache[childrenKey].([]TreeNode), n)
		}
	}
}

func (builder *TreeBuilder) Build() []TreeNode {
	return builder.cache[RootsNodeKey].([]TreeNode)
}
