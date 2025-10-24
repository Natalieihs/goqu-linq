// group.go

package core

import (
	"github.com/doug-martin/goqu/v9"
)

// AggregateInfo 定义聚合操作的信息
type AggregateInfo struct {
	Field    string
	Function string
	Alias    string
}

// GroupAggregateBuilder 用于构建分组聚合查询
type GroupAggregateBuilder[T any] struct {
	aggregations []AggregateInfo
}

type GroupingQuery[T any] struct {
	parent      *Queryable[T]
	keySelector func(T) interface{}
}

// NewGroupAggregateBuilder 创建分组聚合构建器
func NewGroupAggregateBuilder[T any]() *GroupAggregateBuilder[T] {
	return &GroupAggregateBuilder[T]{
		aggregations: make([]AggregateInfo, 0),
	}
}

// Sum 添加求和聚合
func (b *GroupAggregateBuilder[T]) Sum(field string) *GroupAggregateBuilder[T] {
	b.aggregations = append(b.aggregations, AggregateInfo{
		Field:    field,
		Function: "SUM",
	})
	return b
}

// Average 添加平均值聚合
func (b *GroupAggregateBuilder[T]) Average(field string) *GroupAggregateBuilder[T] {
	b.aggregations = append(b.aggregations, AggregateInfo{
		Field:    field,
		Function: "AVG",
	})
	return b
}

// Count 添加计数聚合
func (b *GroupAggregateBuilder[T]) Count() *GroupAggregateBuilder[T] {
	b.aggregations = append(b.aggregations, AggregateInfo{
		Field:    "*",
		Function: "COUNT",
	})
	return b
}

// Max 添加最大值聚合
func (b *GroupAggregateBuilder[T]) Max(field string) *GroupAggregateBuilder[T] {
	b.aggregations = append(b.aggregations, AggregateInfo{
		Field:    field,
		Function: "MAX",
	})
	return b
}

// Min 添加最小值聚合
func (b *GroupAggregateBuilder[T]) Min(field string) *GroupAggregateBuilder[T] {
	b.aggregations = append(b.aggregations, AggregateInfo{
		Field:    field,
		Function: "MIN",
	})
	return b
}

// WithAlias 为最后添加的聚合操作设置别名
func (b *GroupAggregateBuilder[T]) WithAlias(alias string) *GroupAggregateBuilder[T] {
	if len(b.aggregations) > 0 {
		b.aggregations[len(b.aggregations)-1].Alias = alias
	}
	return b
}

// GetAggregations 获取所有聚合操作
func (b *GroupAggregateBuilder[T]) GetAggregations() []AggregateInfo {
	return b.aggregations
}

func (g *GroupingQuery[T]) Select(selector func(key interface{}, elements []T) interface{}) IQueryable[T] {
	selects := []interface{}{
		g.keySelector, // 分组键
	}
	g.parent.query = g.parent.query.Select(selects...).GroupBy(g.keySelector)
	return g.parent
}

func (g *GroupingQuery[T]) Count() (map[interface{}]int64, error) {
	// 构造 SELECT 子句，包含分组键和计数
	selects := []interface{}{
		g.keySelector,               // 分组键
		goqu.COUNT("*").As("count"), // 计数
	}

	g.parent.query = g.parent.query.Select(selects...).GroupBy(g.keySelector)

	// 执行查询
	rows, err := g.parent.db.DB.Queryx(g.parent.query.ToSQL())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 处理结果
	results := make(map[interface{}]int64)
	for rows.Next() {
		var key interface{}
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			return nil, err
		}
		results[key] = count
	}

	return results, nil
}

func (g *GroupingQuery[T]) Sum(field string) (map[interface{}]float64, error) {
	selects := []interface{}{
		g.keySelector,
		goqu.SUM(field).As("sum"),
	}

	g.parent.query = g.parent.query.Select(selects...).GroupBy(g.keySelector)
	sql, args, err := g.parent.query.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := g.parent.db.DB.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[interface{}]float64)
	for rows.Next() {
		var key interface{}
		var sum float64
		if err := rows.Scan(&key, &sum); err != nil {
			return nil, err
		}
		results[key] = sum
	}

	return results, nil
}

func (g *GroupingQuery[T]) Average(field string) (map[interface{}]float64, error) {
	selects := []interface{}{
		g.keySelector,
		goqu.AVG(field).As("avg"),
	}

	g.parent.query = g.parent.query.Select(selects...).GroupBy(g.keySelector)
	sql, args, err := g.parent.query.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := g.parent.db.DB.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[interface{}]float64)
	for rows.Next() {
		var key interface{}
		var avg float64
		if err := rows.Scan(&key, &avg); err != nil {
			return nil, err
		}
		results[key] = avg
	}

	return results, nil
}

// 在 group.go 中添加 Having 方法的实现
func (g *GroupingQuery[T]) Having(condition goqu.Ex) IGroupingQuery[T] {
	g.parent.query = g.parent.query.Having(condition)
	return g
}

func (g *GroupingQuery[T]) Aggregate(builder *GroupAggregateBuilder[T]) IQueryable[T] {
	selects := make([]interface{}, 0)

	// 添加分组字段
	selects = append(selects, g.keySelector)

	// 添加聚合表达式
	for _, agg := range builder.GetAggregations() {
		switch agg.Function {
		case "SUM":
			if agg.Alias != "" {
				selects = append(selects, goqu.SUM(agg.Field).As(agg.Alias))
			} else {
				selects = append(selects, goqu.SUM(agg.Field))
			}
		case "AVG":
			if agg.Alias != "" {
				selects = append(selects, goqu.AVG(agg.Field).As(agg.Alias))
			} else {
				selects = append(selects, goqu.AVG(agg.Field))
			}
		case "COUNT":
			if agg.Alias != "" {
				selects = append(selects, goqu.COUNT(agg.Field).As(agg.Alias))
			} else {
				selects = append(selects, goqu.COUNT(agg.Field))
			}
		case "MAX":
			if agg.Alias != "" {
				selects = append(selects, goqu.MAX(agg.Field).As(agg.Alias))
			} else {
				selects = append(selects, goqu.MAX(agg.Field))
			}
		case "MIN":
			if agg.Alias != "" {
				selects = append(selects, goqu.MIN(agg.Field).As(agg.Alias))
			} else {
				selects = append(selects, goqu.MIN(agg.Field))
			}
		}
	}

	g.parent.query = g.parent.query.Select(selects...).GroupBy(g.keySelector)
	return g.parent
}
