package core

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type Queryable[T any] struct {
	db    *DBLogger
	query *goqu.SelectDataset
}

func (q *Queryable[T]) Where(condition goqu.Ex) IQueryable[T] {
	q.query = q.query.Where(condition)
	return q
}

func (q *Queryable[T]) WhereRaw(condition string, args ...interface{}) IQueryable[T] {
	q.query = q.query.Where(goqu.L(condition, args...))
	return q
}

func (q *Queryable[T]) OrderBy(cols ...string) IQueryable[T] {
	orderedExpressions := make([]exp.OrderedExpression, len(cols))
	for i, col := range cols {
		orderedExpressions[i] = goqu.I(col).Asc()
	}
	q.query = q.query.Order(orderedExpressions...)
	return q
}

// OrderByRaw 支持原始排序语句
// OrderByRaw 支持原始排序语句
func (q *Queryable[T]) OrderByRaw(column string) IQueryable[T] {
	// 多字段（含逗号）处理逻辑
	if strings.Contains(column, ",") {
		orderParts := strings.Split(column, ",")
		orderedExpressions := make([]exp.OrderedExpression, 0, len(orderParts))

		for _, part := range orderParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			fields := strings.Fields(part) // e.g. "id desc" -> ["id", "desc"]
			col := fields[0]
			dir := "ASC"
			if len(fields) > 1 {
				dir = strings.ToUpper(fields[1])
			}

			if dir == "DESC" {
				orderedExpressions = append(orderedExpressions, goqu.I(col).Desc())
			} else {
				orderedExpressions = append(orderedExpressions, goqu.I(col).Asc())
			}
		}

		if len(orderedExpressions) > 0 {
			q.query = q.query.Order(orderedExpressions...)
		}
		return q
	}

	part := strings.TrimSpace(column)
	if part == "" {
		return q
	}

	fields := strings.Fields(part) // e.g. "id desc" -> ["id", "desc"]
	col := fields[0]
	dir := "ASC"
	if len(fields) > 1 {
		dir = strings.ToUpper(fields[1])
	}

	if dir == "DESC" {
		q.query = q.query.Order(goqu.I(col).Desc())
	} else {
		q.query = q.query.Order(goqu.I(col).Asc())
	}
	return q
}

func (q *Queryable[T]) Skip(offset int) IQueryable[T] {
	q.query = q.query.Offset(uint(offset))
	return q
}

func (q *Queryable[T]) Take(limit int) IQueryable[T] {
	q.query = q.query.Limit(uint(limit))
	return q
}

// 需要添加的方法
func (q *Queryable[T]) FirstOrDefault() (*T, error) {
	// 🔥 优化：确保使用结构体字段
	q.ensureSelectFields()

	query, args, err := q.query.Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = q.db.Get(&result, query, args...)
	return &result, err
}

// FirstOrDefaultTx(ctx context.Context) (*T, error)
func (q *Queryable[T]) FirstOrDefaultTx(ctx context.Context) (*T, error) {
	// 🔥 优化：确保使用结构体字段
	q.ensureSelectFields()

	query, args, err := q.query.Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = q.db.GetContext(ctx, &result, query, args...)
	return &result, err
}
func (q *Queryable[T]) ToListTx(ctx context.Context) ([]*T, error) {
	// 🔥 优化：确保使用结构体字段
	q.ensureSelectFields()

	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []*T
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) CountTx(ctx context.Context) (int64, error) {
	query, args, err := q.query.Select(goqu.COUNT("*")).ToSQL()
	if err != nil {
		return 0, err
	}
	var count int64
	err = q.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// sum
func (q *Queryable[T]) SumTx(ctx context.Context, field string) (float64, error) {
	query, args, err := q.query.Select(goqu.SUM(field)).ToSQL()
	if err != nil {
		return 0, err
	}
	var sum float64
	err = q.db.GetContext(ctx, &sum, query, args...)
	return sum, err
}

func (q *Queryable[T]) ToGroupedListTx(ctx context.Context) ([]*T, error) {
	// 🔥 优化：确保使用结构体字段
	q.ensureSelectFields()

	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []*T
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) AnyTx(ctx context.Context, condition goqu.Ex) (bool, error) {
	count, err := q.Where(condition).CountTx(ctx)
	return count > 0, err
}

// Min
func (q *Queryable[T]) MinTx(ctx context.Context, field string) (interface{}, error) {
	query, args, err := q.query.Select(goqu.MIN(field)).ToSQL()
	if err != nil {
		return nil, err
	}
	var min interface{}
	err = q.db.GetContext(ctx, &min, query, args...)
	return min, err
}

func (q *Queryable[T]) ToPagedListTx(ctx context.Context, page, size int, condition goqu.Ex) (*PageResult[T], error) {
	offset := (page - 1) * size
	items, err := q.Where(condition).Skip(offset).Take(size).ToListTx(ctx)
	if err != nil {
		return nil, err
	}
	count, err := q.Where(condition).CountTx(ctx)
	if err != nil {
		return nil, err
	}
	return &PageResult[T]{
		Items:    items,
		Total:    count,
		Page:     page,
		PageSize: size,
	}, nil
}

func (q *Queryable[T]) ToInt64SliceTx(ctx context.Context) ([]int64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []int64
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToStringSliceTx(ctx context.Context) ([]string, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []string
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToFloat64SliceTx(ctx context.Context) ([]float64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []float64
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToMapSliceTx(ctx context.Context) ([]map[string]interface{}, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []map[string]interface{}
	err = q.db.SelectContext(ctx, &results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToMapTx(ctx context.Context) (map[string]interface{}, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = q.db.GetContext(ctx, &result, query, args...)
	return result, err
}

func (q *Queryable[T]) ToStructTx(ctx context.Context) (*T, error) {
	query, args, err := q.query.Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = q.db.GetContext(ctx, &result, query, args...)
	return &result, err
}

func (q *Queryable[T]) ToResultTx(ctx context.Context, result interface{}) error {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return err
	}
	return q.db.SelectContext(ctx, result, query, args...)
}

// MaxTx
func (q *Queryable[T]) MaxTx(ctx context.Context, field string) (interface{}, error) {
	query, args, err := q.query.Select(goqu.MAX(field)).ToSQL()
	if err != nil {
		return nil, err
	}
	var max interface{}
	err = q.db.GetContext(ctx, &max, query, args...)
	return max, err
}

func (q *Queryable[T]) ToList() ([]*T, error) {
	// 🔥 优化：如果没有指定 Select 字段，自动使用结构体中定义的字段
	q.ensureSelectFields()

	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []*T
	err = q.db.Select(&results, query, args...)
	return results, err
}

func (q *Queryable[T]) Count() (int64, error) {
	query, args, err := q.query.Select(goqu.COUNT("*")).ToSQL()
	if err != nil {
		return 0, err
	}
	var count int64
	err = q.db.Get(&count, query, args...)
	return count, err
}

func (q *Queryable[T]) ToGroupedList() ([]*T, error) {
	// 🔥 优化：确保使用结构体字段
	q.ensureSelectFields()

	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []*T
	err = q.db.Select(&results, query, args...)
	return results, err
}
func (q *Queryable[T]) Any(condition goqu.Ex) (bool, error) {
	count, err := q.Where(condition).Count()
	return count > 0, err
}

func (q *Queryable[T]) Sum(field string) (float64, error) {
	sumExpr := goqu.L("IFNULL(SUM(?), 0)", goqu.I(field))
	query, args, err := q.query.Select(sumExpr).ToSQL()
	if err != nil {
		return 0, err
	}
	var sum float64
	err = q.db.Get(&sum, query, args...)
	return sum, err
}
func (q *Queryable[T]) Max(field string) (interface{}, error) {
	query, args, err := q.query.Select(goqu.MAX(field)).ToSQL()
	if err != nil {
		return nil, err
	}
	var max interface{}
	err = q.db.Get(&max, query, args...)
	return max, err
}

// 在 Queryable 中添加
func (q *Queryable[T]) ToPagedList(page, size int, condition goqu.Ex) (*PageResult[T], error) {
	offset := (page - 1) * size
	items, err := q.Where(condition).Skip(offset).Take(size).ToList()
	if err != nil {
		return nil, err
	}

	total, err := q.Count()
	if err != nil {
		return nil, err
	}
	return &PageResult[T]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: size,
	}, nil
}

// getSelectColumns 根据结构体的 db tag 自动生成 SELECT 列
func getSelectColumns[T any]() []interface{} {
	var entity T
	t := reflect.TypeOf(entity)

	// 处理指针类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []interface{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")

		// 跳过没有 db tag 或标记为 "-" 的字段
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// 使用 db tag 的值作为列名
		columns = append(columns, goqu.I(dbTag))
	}

	// 如果没有找到任何列，返回 * 作为兜底
	if len(columns) == 0 {
		return []interface{}{goqu.Star()}
	}

	return columns
}

// Note: Query() method is now defined in repository.go

// Select(cols ...interface{}) IQueryable[T]
func (q *Queryable[T]) Select(cols ...interface{}) IQueryable[T] {
	q.query = q.query.Select(cols...)
	return q
}

func (q *Queryable[T]) SelectRaw(cols ...string) IQueryable[T] {
	expressions := make([]interface{}, len(cols))
	for i, col := range cols {
		expressions[i] = goqu.L(col)
	}
	q.query = q.query.Select(expressions...)
	return q
}

// Min(field string) (interface{}, error)
func (q *Queryable[T]) Min(field string) (interface{}, error) {
	query, args, err := q.query.Select(goqu.MIN(field)).ToSQL()
	if err != nil {
		return nil, err
	}
	var min interface{}
	err = q.db.Get(&min, query, args...)
	return min, err
}

// //泛型方法 转slice
// ToInt64Slice() ([]int64, error)
// ToStringSlice() ([]string, error)
// ToFloat64Slice() ([]float64, error)
// ToMapSlice() ([]map[string]interface{}, error)
// ToMap() (map[string]interface{}, error)
// //泛型方法 转struct
// ToStruct() (*T, error)

// 在 Queryable 中添加
func (q *Queryable[T]) ToInt64Slice() ([]int64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []int64
	err = q.db.Select(&results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToStringSlice() ([]string, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []string
	err = q.db.Select(&results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToFloat64Slice() ([]float64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []float64
	err = q.db.Select(&results, query, args...)
	return results, err
}

func (q *Queryable[T]) ToMapSlice() ([]map[string]interface{}, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}

	// 1. 先扫描到结构体切片
	var items []*T
	err = q.db.Select(&items, query, args...)
	if err != nil {
		return nil, err
	}

	// 2. 结构体切片转为 map 切片
	results := make([]map[string]interface{}, len(items))
	for i, item := range items {
		// 使用反射将结构体转为 map
		v := reflect.ValueOf(item).Elem()
		t := v.Type()
		m := make(map[string]interface{})

		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			// 获取 db tag
			if tag := field.Tag.Get("db"); tag != "" {
				m[tag] = v.Field(j).Interface()
			}
		}
		results[i] = m
	}

	return results, nil
}

func (q *Queryable[T]) ToMap() (map[string]interface{}, error) {
	//用rows.Next() 的方式
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := q.db.DB.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	// 获取列类型
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	// 创建一个切片用于存储列值
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}
	// 创建一个 map 用于存储结果
	result := make(map[string]interface{})
	// 遍历行
	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		// 遍历列
		for i, column := range columns {
			// 获取列类型
			t := types[i]
			// 获取列值
			value := *(values[i].(*interface{}))
			// 转换列值
			switch t.DatabaseTypeName() {
			case "INT":
				result[column] = value
			case "VARCHAR":
				result[column] = value
			case "FLOAT":
				result[column] = value
			}
		}
	}
	return result, nil
}
func (q *Queryable[T]) ToStruct() (*T, error) {
	query, args, err := q.query.Limit(1).ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = q.db.Get(&result, query, args...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Limit(limit int) IQueryable[T]
func (q *Queryable[T]) Limit(limit int) IQueryable[T] {
	q.query = q.query.Limit(uint(limit))
	return q
}

// 写一个 Scan(&maxSort) 的方法
func (q *Queryable[T]) Scan(dest interface{}) error {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return err
	}
	return q.db.Get(dest, query, args...)
}

// ScanTx(ctx context.Context, dest interface{}) error
func (q *Queryable[T]) ScanTx(ctx context.Context, dest interface{}) error {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return err
	}
	return q.db.GetContext(ctx, dest, query, args...)
}

/*
ScanInt64() (int64, error)

	ScanString() (string, error)
	ScanInt() (int, error)
*/
func (q *Queryable[T]) ScanInt64() (int64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return 0, err
	}
	var result int64
	err = q.db.Get(&result, query, args...)
	return result, err
}

func (q *Queryable[T]) ScanString() (string, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return "", err
	}
	var result string
	err = q.db.Get(&result, query, args...)
	return result, err
}

func (q *Queryable[T]) ScanInt() (int, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return 0, err
	}
	var result int
	err = q.db.Get(&result, query, args...)
	return result, err
}

// ScanVal() (interface{}, error)
func (q *Queryable[T]) ScanVal() (interface{}, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = q.db.Get(&result, query, args...)
	return result, err
}

// ToLookup(keySelector func(T) interface{}) map[interface{}][]*T
func (q *Queryable[T]) ToLookup(keySelector func(T) interface{}) map[interface{}][]*T {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil
	}
	var results []*T
	err = q.db.Select(&results, query, args...)
	if err != nil {
		return nil
	}
	lookup := make(map[interface{}][]*T)
	for _, item := range results {
		key := keySelector(*item)
		lookup[key] = append(lookup[key], item)
	}
	return lookup
}

// ToPagedListWithTotal(page, size int, condition goqu.Ex) ([]*T, int64, error)
func (q *Queryable[T]) ToPagedListWithTotal(page, size int, condition goqu.Ex) ([]*T, int64, error) {
	//先查询总数
	total, err := q.Where(condition).Count()
	if err != nil {
		return nil, 0, err
	}
	//再查询分页数据
	offset := (page - 1) * size
	items, err := q.Where(condition).Skip(offset).Take(size).ToList()
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// 支持窗口函数，如 ROW_NUMBER, RANK, DENSE_RANK 等
func (q *Queryable[T]) Over(windowFunc string, partitionBy ...interface{}) IQueryable[T] {
	w := goqu.W()
	if len(partitionBy) > 0 {
		w = w.PartitionBy(partitionBy...)
	}
	q.query = q.query.Select(
		goqu.L(fmt.Sprintf("%s() OVER ?", windowFunc), w),
	)
	return q
}

func (q *Queryable[T]) GroupByHaving(having goqu.Ex) IQueryable[T] {
	q.query = q.query.Having(having)
	return q
}

func (q *Queryable[T]) ToResult(result interface{}) error {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return err
	}
	return q.db.Select(result, query, args...)
}

func (q *Queryable[T]) Join(table string, on map[string]string) IQueryable[T] {
	conditions := make(goqu.Ex)
	for sourceKey, targetKey := range on {
		conditions[sourceKey] = goqu.I(targetKey)
	}
	q.query = q.query.LeftJoin(goqu.T(table), goqu.On(conditions))
	return q
}

func (q *Queryable[T]) LeftJoin(table string, on map[string]string) IQueryable[T] {
	conditions := make(goqu.Ex)
	for sourceKey, targetKey := range on {
		conditions[sourceKey] = goqu.I(targetKey)
	}
	q.query = q.query.LeftJoin(goqu.T(table), goqu.On(conditions))
	return q
}

// RightJoin 实现
func (q *Queryable[T]) RightJoin(table string, on map[string]string) IQueryable[T] {
	conditions := make(goqu.Ex)
	for sourceKey, targetKey := range on {
		conditions[sourceKey] = goqu.I(targetKey)
	}
	q.query = q.query.RightJoin(goqu.T(table), goqu.On(conditions))
	return q
}

// InnerJoin 实现
func (q *Queryable[T]) InnerJoin(table string, on map[string]string) IQueryable[T] {
	conditions := make(goqu.Ex)
	for sourceKey, targetKey := range on {
		conditions[sourceKey] = goqu.I(targetKey)
	}
	q.query = q.query.InnerJoin(goqu.T(table), goqu.On(conditions))
	return q
}

// Queryable 实现
func (q *Queryable[T]) ToPagedResult(page, pageSize int, dest interface{}) (*PagedResult, error) {
	// 1. 获取总记录数
	total, err := q.Count()
	if err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 2. 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// 3. 获取分页数据
	offset := (page - 1) * pageSize
	query := q.query.Offset(uint(offset)).Limit(uint(pageSize))

	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql error: %w", err)
	}

	// 4. 执行查询并填充结果
	err = q.db.Select(dest, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &PagedResult{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// AggregateResult 用于存储聚合结果
// AggregateResult 优化结构
type AggregateResult struct {
	Groups map[string]interface{} // 分组字段的值
	Sums   map[string]float64     // 各字段的sum结果
	Count  int64                  // 该组的记录数
}

// GroupField 结构用于定义分组字段
type GroupField struct {
	Field string
	Alias string
}

// GroupSumMultiple 实现
func (q *Queryable[T]) GroupSumMultiple(groupFields []GroupField, sumFields []string) ([]*AggregateResult, error) {
	// 构造 SELECT 子句
	selects := make([]interface{}, 0, len(groupFields)+len(sumFields))

	// 构造分组字段表达式和映射
	groupExps := make([]interface{}, len(groupFields))
	for i, field := range groupFields {
		if field.Alias != "" {
			selects = append(selects, goqu.I(field.Field).As(field.Alias))
			groupExps[i] = goqu.I(field.Field)
		} else {
			selects = append(selects, field.Field)
			groupExps[i] = field.Field
		}
	}

	// 添加聚合字段
	for _, field := range sumFields {
		selects = append(selects, goqu.SUM(field).As(field+"_sum"))
	}

	// 添加计数
	selects = append(selects, goqu.COUNT("*").As("group_count"))

	// 构造查询
	query := q.query.
		Select(selects...).
		GroupBy(groupExps...)

	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql error: %w", err)
	}

	// 执行查询
	rows, err := q.db.DB.Queryx(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	// 处理结果
	var results []*AggregateResult
	for rows.Next() {
		rowMap := make(map[string]interface{})
		err = rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		result := &AggregateResult{
			Groups: make(map[string]interface{}),
			Sums:   make(map[string]float64),
		}

		// 处理分组字段
		for _, field := range groupFields {
			key := field.Alias
			if key == "" {
				key = field.Field
			}
			if val, ok := rowMap[key]; ok {
				result.Groups[key] = val
			}
		}

		// 处理聚合字段
		for _, field := range sumFields {
			sumKey := field + "_sum"
			if val, ok := rowMap[sumKey]; ok {
				switch v := val.(type) {
				case float64:
					result.Sums[field] = v
				case int64:
					result.Sums[field] = float64(v)
				case nil:
					result.Sums[field] = 0
				}
			}
		}

		// 处理计数
		if count, ok := rowMap["group_count"]; ok {
			if v, ok := count.(int64); ok {
				result.Count = v
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (q *Queryable[T]) GroupBy(keySelector func(T) interface{}) IGroupingQuery[T] {
	return &GroupingQuery[T]{
		parent:      q,
		keySelector: keySelector,
	}
}

// 在 Queryable 中添加 GroupByColumns 实现
func (q *Queryable[T]) GroupByColumns(cols ...string) IQueryable[T] {
	colsInterface := make([]interface{}, len(cols))
	for i, col := range cols {
		colsInterface[i] = col
	}
	q.query = q.query.GroupBy(colsInterface...)
	return q
}

// OverTx
func (q *Queryable[T]) OverTx(ctx context.Context, windowFunc string, partitionBy ...interface{}) IQueryable[T] {
	w := goqu.W()
	if len(partitionBy) > 0 {
		w = w.PartitionBy(partitionBy...)
	}
	q.query = q.query.Select(
		goqu.L(fmt.Sprintf("%s() OVER ?", windowFunc), w),
	)
	return q
}

// ToLookupTx
func (q *Queryable[T]) ToLookupTx(ctx context.Context, keySelector func(T) interface{}) map[interface{}][]*T {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil
	}
	var results []*T
	err = q.db.SelectContext(ctx, &results, query, args...)
	if err != nil {
		return nil
	}
	lookup := make(map[interface{}][]*T)
	for _, item := range results {
		key := keySelector(*item)
		lookup[key] = append(lookup[key], item)
	}
	return lookup
}

// ToPagedResultTx
func (q *Queryable[T]) ToPagedResultTx(ctx context.Context, page, pageSize int, dest interface{}) (*PagedResult, error) {
	// 1. 获取总记录数
	total, err := q.CountTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("count error: %w", err)
	}

	// 2. 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// 3. 获取分页数据
	offset := (page - 1) * pageSize
	query := q.query.Offset(uint(offset)).Limit(uint(pageSize))

	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("build sql error: %w", err)
	}

	// 4. 执行查询并填充结果
	err = q.db.SelectContext(ctx, dest, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}

	return &PagedResult{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ScanInt64Slice() ([]int64, error)
func (q *Queryable[T]) ScanInt64Slice() ([]int64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return nil, err
	}
	var results []int64
	err = q.db.Select(&results, query, args...)
	return results, err
}

// ScanFloat64() (float64, error)
func (q *Queryable[T]) ScanFloat64() (float64, error) {
	query, args, err := q.query.ToSQL()
	if err != nil {
		return 0, err
	}
	var result float64
	err = q.db.Get(&result, query, args...)
	return result, err
}

func (q *Queryable[T]) ToSQL() (sql string, params []interface{}, err error) {
	query, args, err := q.query.ToSQL()
	return query, args, err
}

// ensureSelectFields 确保查询中包含 SELECT 字段
// 如果没有指定 Select，则自动使用结构体中定义的字段
func (q *Queryable[T]) ensureSelectFields() {
	// 检查是否已经有 SELECT 子句
	sql, _, _ := q.query.ToSQL()

	// 如果 SQL 包含 "SELECT  FROM" (两个空格) 说明没有指定字段
	if strings.Contains(sql, "SELECT  FROM") {
		// 获取结构体的所有数据库字段
		fields := q.getStructDBFields()
		if len(fields) > 0 {
			q.query = q.query.Select(fields...)
		}
	}
}

// getStructDBFields 获取结构体中定义的所有数据库字段
func (q *Queryable[T]) getStructDBFields() []interface{} {
	var t T
	typ := reflect.TypeOf(t)

	// 如果是指针类型，获取其指向的类型
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// 只处理结构体类型
	if typ.Kind() != reflect.Struct {
		return nil
	}

	var fields []interface{}

	// 遍历结构体的所有字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 获取 db tag
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			// 只取 tag 的第一部分（逗号之前）
			dbFieldName := strings.Split(dbTag, ",")[0]
			fields = append(fields, dbFieldName)
		}
	}

	return fields
}
