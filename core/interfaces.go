package core

// contrib/dao/base/interfaces.go
import (
	"context"

	"github.com/doug-martin/goqu/v9"
)

// contrib/dao/base/interfaces.go
type IReadRepository[T any] interface {
	Query() IQueryable[T]
	// Where(condition goqu.Ex) IQueryable[T]
	// OrderBy(cols ...string) IQueryable[T]
	// Skip(offset int) IQueryable[T]
	// Take(limit int) IQueryable[T]
	// FirstOrDefault() (*T, error)
	// ToList() ([]*T, error)
	// Count() (int64, error)
	// Select(cols ...interface{}) IRepository[T]
	// GroupBy(cols ...string) IRepository[T]
	// ToGroupedList() ([]*T, error)
	// ToPagedList(page, size int, condition goqu.Ex) (*PageResult[T], error)
	// Any(condition goqu.Ex) (bool, error)
	// Sum(field string) (float64, error)
	// Max(field string) (interface{}, error)
	// Min(field string) (interface{}, error)

	// //Over() IRepository[T]
	// //泛型方法 转slice
	// ToInt64Slice() ([]int64, error)
	// ToStringSlice() ([]string, error)
	// ToFloat64Slice() ([]float64, error)
	// ToMapSlice() ([]map[string]interface{}, error)
	// ToMap() (map[string]interface{}, error)
	// //泛型方法 转struct
	// ToStruct() (*T, error)
}

type IWriteRepository[T any] interface {
	Create(entity *T) error
	Update(entity *T) error
	UpdateByCondition(condition goqu.Ex, entity *T) error
	UpdateFieldsByCondition(condition goqu.Ex, fields map[string]interface{}) error
	BatchCreate(entities []*T) error
	BatchDelete(condition goqu.Ex) error
	BatchInsert(entities []*T, opt *BatchInsertOption) error
	BatchUpdate(entities []*T, opt *BatchUpdateOption) error
}

type IRepository[T any] interface {
	IReadRepository[T]
	IWriteRepository[T]
}

// IQueryable 接口增加 Lambda 风格的分组方法
type IQueryable[T any] interface {
	// 现有的链式操作
	Where(condition goqu.Ex) IQueryable[T]
	WhereRaw(condition string, args ...interface{}) IQueryable[T]
	OrderBy(cols ...string) IQueryable[T]
	OrderByRaw(order string) IQueryable[T]
	Skip(offset int) IQueryable[T]
	Take(take int) IQueryable[T]
	Limit(limit int) IQueryable[T]
	Scan(dest interface{}) error
	ScanTx(ctx context.Context, dest interface{}) error
	ScanVal() (interface{}, error)
	ScanInt64() (int64, error)
	ScanFloat64() (float64, error)
	ScanString() (string, error)
	ScanInt64Slice() ([]int64, error)
	ScanInt() (int, error)
	//新增一个泛型方法 scan

	Select(cols ...interface{}) IQueryable[T]
	SelectRaw(cols ...string) IQueryable[T] // 原生 SQL 查询
	// 分组操作 - 新增 Lambda 风格
	GroupBy(keySelector func(T) interface{}) IGroupingQuery[T]
	// 保留原有的字符串方式，用于简单场景
	GroupByColumns(cols ...string) IQueryable[T]

	// 连接操作
	IJoinable[T]

	// 执行方法
	FirstOrDefault() (*T, error)

	//FirstOrDefaultTx(ctx context.Context) (*T, error)

	ToListTx(ctx context.Context) ([]*T, error)

	CountTx(ctx context.Context) (int64, error)

	ToGroupedListTx(ctx context.Context) ([]*T, error)
	AnyTx(ctx context.Context, condition goqu.Ex) (bool, error)
	SumTx(ctx context.Context, field string) (float64, error)
	MaxTx(ctx context.Context, field string) (interface{}, error)
	MinTx(ctx context.Context, field string) (interface{}, error)
	ToPagedListTx(ctx context.Context, page, size int, condition goqu.Ex) (*PageResult[T], error)
	ToInt64SliceTx(ctx context.Context) ([]int64, error)
	ToStringSliceTx(ctx context.Context) ([]string, error)
	ToFloat64SliceTx(ctx context.Context) ([]float64, error)
	ToMapSliceTx(ctx context.Context) ([]map[string]interface{}, error)
	ToMapTx(ctx context.Context) (map[string]interface{}, error)
	ToStructTx(ctx context.Context) (*T, error)
	ToResultTx(ctx context.Context, result interface{}) error
	ToPagedResultTx(ctx context.Context, page, pageSize int, dest interface{}) (*PagedResult, error)
	OverTx(ctx context.Context, windowFunc string, partitionBy ...interface{}) IQueryable[T]
	ToLookupTx(ctx context.Context, keySelector func(T) interface{}) map[interface{}][]*T
	ToList() ([]*T, error)
	Count() (int64, error)
	Any(condition goqu.Ex) (bool, error)

	// 聚合方法
	Sum(field string) (float64, error)
	Max(field string) (interface{}, error)
	Min(field string) (interface{}, error)

	// 分页相关
	ToPagedList(page, size int, condition goqu.Ex) (*PageResult[T], error)
	ToPagedListWithTotal(page, size int, condition goqu.Ex) ([]*T, int64, error)
	ToPagedResult(page, pageSize int, dest interface{}) (*PagedResult, error)

	// 结果转换
	ToInt64Slice() ([]int64, error)
	ToStringSlice() ([]string, error)
	ToFloat64Slice() ([]float64, error)
	ToMapSlice() ([]map[string]interface{}, error)
	ToMap() (map[string]interface{}, error)
	ToStruct() (*T, error)
	ToResult(result interface{}) error

	// 特殊查询
	Over(windowFunc string, partitionBy ...interface{}) IQueryable[T]
	ToLookup(keySelector func(T) interface{}) map[interface{}][]*T
	ToSQL() (sql string, params []interface{}, err error)
}

// IJoinable 接口定义简化的连表操作
type IJoinable[T any] interface {
	Join(table string, on map[string]string) IQueryable[T]
	LeftJoin(table string, on map[string]string) IQueryable[T]
	RightJoin(table string, on map[string]string) IQueryable[T]
	InnerJoin(table string, on map[string]string) IQueryable[T]
}

type PagedResult struct {
	Total      int64 // 总记录数
	Page       int   // 当前页码
	PageSize   int   // 每页大小
	TotalPages int   // 总页数
}

type IGrouping[TKey any, TElement any] interface {
	Key() TKey
	Elements() []TElement
}

type GroupResult[TKey any, TElement any] struct {
	key      TKey
	elements []TElement
}

func (g *GroupResult[TKey, TElement]) Key() TKey {
	return g.key
}

func (g *GroupResult[TKey, TElement]) Elements() []TElement {
	return g.elements
}

// IGroupingQuery 分组查询接口
type IGroupingQuery[T any] interface {
	// 基础聚合操作
	Count() (map[interface{}]int64, error)
	Sum(field string) (map[interface{}]float64, error)
	Average(field string) (map[interface{}]float64, error)

	// 高级操作
	Select(selector func(key interface{}, elements []T) interface{}) IQueryable[T]
	Having(condition goqu.Ex) IGroupingQuery[T]

	// 链式聚合操作
	Aggregate(builder *GroupAggregateBuilder[T]) IQueryable[T]
}
