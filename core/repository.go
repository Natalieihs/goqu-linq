package core

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type DialectType string

const (
	MySQL     DialectType = "mysql"
	StarRocks DialectType = "starrocks"
)

type Repository[T any] struct {
	db *DBLogger
	//query *goqu.SelectDataset
	table   string
	dialect goqu.DialectWrapper
	dbType  DialectType
	uow     IUnitOfWork // 工作单元
}

func (r *Repository[T]) WithUnitOfWork(uow IUnitOfWork) *Repository[T] {
	return &Repository[T]{
		db:      r.db,
		table:   r.table,
		dialect: r.dialect,
		dbType:  r.dbType,
		uow:     uow,
	}
}
func (r *Repository[T]) CreateWithTx(entity *T) error {
	query := r.dialect.Insert(r.table).Rows(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// UpdateWithTx(entity)
func (r *Repository[T]) UpdateWithTx(entity *T) error {
	query := r.dialect.Update(r.table).Set(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// UpdateByConditionWithTx
func (r *Repository[T]) UpdateByConditionWithTx(condition goqu.Ex, entity *T) error {
	query := r.dialect.Update(r.table).Set(entity).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// UpdateFieldsByConditionWithTx
func (r *Repository[T]) UpdateFieldsByConditionWithTx(condition goqu.Ex, fields map[string]interface{}) error {
	updateExp := make(map[string]interface{})
	for field, value := range fields {
		if _, ok := value.(exp.LiteralExpression); ok {
			updateExp[field] = value
		} else {
			updateExp[field] = value
		}
	}
	query := r.dialect.Update(r.table).Set(updateExp).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// BatchDeleteWithTx
func (r *Repository[T]) BatchDeleteWithTx(condition goqu.Ex) error {
	query := r.dialect.Delete(r.table).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// Query returns a queryable interface for building queries
func (r *Repository[T]) Query() IQueryable[T] {
	return &Queryable[T]{
		db:    r.db,
		query: r.dialect.From(r.table),
	}
}

func (r *Repository[T]) QueryFrom(dbType DialectType) IQueryable[T] {
	return &Queryable[T]{
		db:    r.db,
		query: goqu.Dialect("mysql").From(r.table),
	}
}

// 获取db
func (r *Repository[T]) GetDB() *DBLogger {
	return r.db
}

// 获取前缀
func (r *Repository[T]) GetPrefix() string {
	return r.db.prefix
}

func NewRepository[T any](db *DBLogger, table string, dbType DialectType) *Repository[T] {
	var dialect goqu.DialectWrapper
	switch dbType {
	case StarRocks:
		dialect = goqu.Dialect("mysql")
	default:
		dialect = goqu.Dialect("mysql")
	}
	return &Repository[T]{
		db:      db,
		table:   table,
		dialect: dialect,
		dbType:  dbType,
	}
}

func (r *Repository[T]) Create(entity *T) error {
	// 如果有工作单元，调用 CreateWithTx
	if r.uow != nil {
		return r.CreateWithTx(entity)
	}

	// 否则直接执行 SQL
	query := r.dialect.Insert(r.table).Rows(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

func (r *Repository[T]) Update(entity *T) error {
	if r.uow != nil {
		// 需要实现 UpdateWithTx 方法
		return r.UpdateWithTx(entity)
	}
	query := r.dialect.Update(r.table).Set(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

type PageResult[T any] struct {
	Items    []*T
	Total    int64
	Page     int
	PageSize int
}

// 继续写
func (r *Repository[T]) UpdateByCondition(condition goqu.Ex, entity *T) error {
	if r.uow != nil {
		return r.UpdateByConditionWithTx(condition, entity)
	}
	query := r.dialect.Update(r.table).Set(entity).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

// 更新指定字段 根据指定条件
func (r *Repository[T]) UpdateFieldsByCondition(condition goqu.Ex, fields map[string]interface{}) error {
	if r.uow != nil {
		return r.UpdateFieldsByConditionWithTx(condition, fields)
	}
	updateExp := make(map[string]interface{})

	for field, value := range fields {
		// 使用 exp.LiteralExpression 接口类型来判断
		if _, ok := value.(exp.LiteralExpression); ok {
			updateExp[field] = value
		} else {
			updateExp[field] = value
		}
	}

	query := r.dialect.Update(r.table).Set(updateExp).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

// BatchCreate - 批量创建

func (r *Repository[T]) BatchCreate(entities []*T) error {
	query := r.dialect.Insert(r.table).Rows(entities)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}

	if r.uow != nil {
		_, err = r.uow.GetTx().Exec(sql, args...)
	} else {
		_, err = r.db.Exec(sql, args...)
	}
	return err
}

// BatchDelete - 批量删除
func (r *Repository[T]) BatchDelete(condition goqu.Ex) error {
	if r.uow != nil {
		return r.BatchDeleteWithTx(condition)
	}
	query := r.dialect.Delete(r.table).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

// BatchInsertOption 批量插入的配置选项
type BatchInsertOption struct {
	BatchSize    int  // 每批次处理的数据量
	UseNamedExec bool // 是否使用NamedExec方式
}

// DefaultBatchInsertOption 默认的批量插入配置
var DefaultBatchInsertOption = &BatchInsertOption{
	BatchSize:    1000, // 默认每批次1000条
	UseNamedExec: false,
}

// BatchInsert 批量插入数据的通用方法
func (r *Repository[T]) BatchInsert(entities []*T, opt *BatchInsertOption) error {
	if len(entities) == 0 {
		return nil
	}

	if opt == nil {
		opt = DefaultBatchInsertOption
	}

	// 获取字段数量
	fields := r.getFields(entities[0])
	if len(fields) == 0 {
		return fmt.Errorf("no fields found in entity")
	}

	// 计算安全的批次大小
	safeBatchSize := calculateSafeBatchSize(len(fields), 16384)
	if opt.BatchSize > safeBatchSize {
		opt.BatchSize = safeBatchSize
	}

	// 分批处理
	for i := 0; i < len(entities); i += opt.BatchSize {
		end := i + opt.BatchSize
		if end > len(entities) {
			end = len(entities)
		}

		batch := entities[i:end]
		if opt.UseNamedExec {
			if err := r.batchInsertByNamedExec(batch); err != nil {
				return fmt.Errorf("batch insert failed at offset %d: %w", i, err)
			}
		} else {
			if err := r.batchInsertByExec(batch); err != nil {
				return fmt.Errorf("batch insert failed at offset %d: %w", i, err)
			}
		}
	}

	return nil
}

// batchInsertByExec 使用手动拼接SQL的方式批量插入
func (r *Repository[T]) batchInsertByExec(entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	fields := r.getFields(entities[0])
	if len(fields) == 0 {
		return fmt.Errorf("no fields found in entity")
	}

	// 构造占位符
	placeholderItems := make([]string, len(fields))
	for i := range fields {
		placeholderItems[i] = "?"
	}
	placeholder := fmt.Sprintf("(%s)", strings.Join(placeholderItems, ","))

	placeholders := make([]string, len(entities))
	for i := range entities {
		placeholders[i] = placeholder
	}

	// 构造SQL
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		r.table,
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	// 准备参数
	values := make([]interface{}, 0, len(entities)*len(fields))
	for _, entity := range entities {
		vals := r.getValues(entity)
		values = append(values, vals...)
	}

	// 执行SQL
	_, err := r.db.Exec(query, values...)
	return err
}

// 计算安全的批次大小
func calculateSafeBatchSize(fieldCount int, maxParams int) int {
	// MySQL 默认 max_prepared_stmt_count 是 16384
	// 每条记录占用 fieldCount 个参数
	// 为安全起见，取最大参数数的80%
	safeMaxParams := maxParams * 80 / 100
	return safeMaxParams / fieldCount
}

// batchInsertByNamedExec 使用NamedExec的方式批量插入
func (r *Repository[T]) batchInsertByNamedExec(entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	// 获取字段名
	fields := r.getFields(entities[0])

	// 构造命名参数占位符
	placeholders := make([]string, len(fields))
	for i, field := range fields {
		placeholders[i] = ":" + field
	}

	// 构造SQL
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		r.table,
		strings.Join(fields, ","),
		strings.Join(placeholders, ","),
	)

	// 执行带命名参数的SQL
	_, err := r.db.NamedExec(query, entities)
	return err
}

// getFields 获取实体的字段名（需要根据实际的标签或反射实现）
func (r *Repository[T]) getFields(entity *T) []string {
	v := reflect.ValueOf(*entity)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	fields := make([]string, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 获取 db tag
		if tag := field.Tag.Get("db"); tag != "" {
			fields = append(fields, tag)
		}
	}
	return fields
}

// getValues 获取实体的字段值（需要根据实际的标签或反射实现）
func (r *Repository[T]) getValues(entity *T) []interface{} {
	v := reflect.ValueOf(*entity)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	values := make([]interface{}, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("db"); tag != "" {
			values = append(values, v.Field(i).Interface())
		}
	}
	return values
}

// BatchUpdateOption 批量更新的配置选项
type BatchUpdateOption struct {
	BatchSize       int      // 每批次处理的数据量
	UpdateFields    []string // 需要更新的字段
	KeyField        string   // 用于WHERE条件的键字段
	AdditionalWhere goqu.Ex  // 附加的WHERE条件
}

// DefaultBatchUpdateOption 默认的批量更新配置
var DefaultBatchUpdateOption = &BatchUpdateOption{
	BatchSize: 1000, // 默认每批次1000条
}

// BatchUpdate 批量更新数据
// entities 要更新的实体数组
// opt 更新选项
func (r *Repository[T]) BatchUpdate(entities []*T, opt *BatchUpdateOption) error {
	if len(entities) == 0 {
		return nil
	}

	if opt == nil {
		opt = DefaultBatchUpdateOption
	}

	// 如果没有指定更新字段，获取所有字段
	if len(opt.UpdateFields) == 0 {
		opt.UpdateFields = r.getFields(entities[0])
	}

	// 确保有主键字段
	if opt.KeyField == "" {
		return fmt.Errorf("key field must be specified for batch update")
	}

	// 计算安全的批次大小（考虑WHERE IN的限制和参数数量限制）
	safeBatchSize := calculateSafeBatchSize(len(opt.UpdateFields)+1, 16384)
	if opt.BatchSize > safeBatchSize {
		opt.BatchSize = safeBatchSize
	}

	// 分批处理
	for i := 0; i < len(entities); i += opt.BatchSize {
		end := i + opt.BatchSize
		if end > len(entities) {
			end = len(entities)
		}

		batch := entities[i:end]
		if err := r.batchUpdateExec(batch, opt); err != nil {
			return fmt.Errorf("batch update failed at offset %d: %w", i, err)
		}
	}

	return nil
}

// batchUpdateExec 执行批量更新
func (r *Repository[T]) batchUpdateExec(entities []*T, opt *BatchUpdateOption) error {
	if len(entities) == 0 {
		return nil
	}

	// 构建 CASE WHEN 语句
	//var cases []string
	var args []interface{}

	// 收集所有的主键值
	keyValues := make([]interface{}, len(entities))
	keyValueMap := make(map[interface{}]*T)

	for i, entity := range entities {
		v := reflect.ValueOf(*entity)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		keyValue := r.getFieldValue(entity, opt.KeyField)
		keyValues[i] = keyValue
		keyValueMap[keyValue] = entity
	}

	// 构建基础SQL
	baseSQL := fmt.Sprintf("UPDATE %s SET ", r.table)

	// 为每个要更新的字段构建 CASE 语句
	setClauses := make([]string, 0, len(opt.UpdateFields))
	for _, field := range opt.UpdateFields {
		if field == opt.KeyField {
			continue
		}

		caseStmt := fmt.Sprintf("%s = CASE %s ", field, opt.KeyField)
		for keyValue, entity := range keyValueMap {
			value := r.getFieldValue(entity, field)
			caseStmt += fmt.Sprintf("WHEN ? THEN ? ")
			args = append(args, keyValue, value)
		}
		caseStmt += "END"
		setClauses = append(setClauses, caseStmt)
	}

	// 组合完整的SQL语句
	sql := baseSQL + strings.Join(setClauses, ", ") +
		fmt.Sprintf(" WHERE %s IN (%s)",
			opt.KeyField,
			strings.Join(strings.Split(strings.Repeat("?", len(keyValues)), ""), ","))

	// 添加WHERE IN的参数
	args = append(args, keyValues...)

	// 如果有附加条件，添加到WHERE子句
	if opt.AdditionalWhere != nil {
		whereSQL, whereArgs, err := r.dialect.From("").Where(opt.AdditionalWhere).ToSQL()
		if err != nil {
			return err
		}
		sql += " AND " + whereSQL[7:] // 去掉开头的"WHERE "
		args = append(args, whereArgs...)
	}

	// 执行SQL
	_, err := r.db.Exec(sql, args...)
	return err
}

// getFieldValue 获取实体指定字段的值
func (r *Repository[T]) getFieldValue(entity *T, fieldName string) interface{} {
	v := reflect.ValueOf(*entity)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("db"); tag == fieldName {
			return v.Field(i).Interface()
		}
	}
	return nil
}

// ----------------------------------------------------------工作单元----------------------------------------------------------
// IUnitOfWork 工作单元接口
type IUnitOfWork interface {
	GetTx() *Tx
	Begin() error
	Commit() error
	Rollback() error
}

// UnitOfWork 实现
type UnitOfWork struct {
	db *DBLogger
	tx *Tx
}

func NewUnitOfWork(db *DBLogger) *UnitOfWork {
	return &UnitOfWork{
		db: db,
	}
}

// 工作单元方法实现
func (u *UnitOfWork) Begin() error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	fmt.Println("事务已成功开始")
	u.tx = tx
	return nil
}

func (u *UnitOfWork) Commit() error {
	return u.tx.Commit()
}

func (u *UnitOfWork) Rollback() error {
	err := u.tx.Rollback()
	if err != nil {
		fmt.Printf("事务回滚失败: %v\n", err)
	} else {
		fmt.Println("事务已成功回滚")
	}
	return err
}
func (u *UnitOfWork) GetTx() *Tx {
	return u.tx
}

// 辅助函数
func (u *UnitOfWork) RunInTransaction(fn func(IUnitOfWork) error) error {
	if err := u.Begin(); err != nil {
		fmt.Printf("开始事务失败: %v\n", err)
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("捕获到异常: %v, 正在回滚事务\n", r)
			u.Rollback()
			panic(r)
		}
	}()

	if err := fn(u); err != nil {
		fmt.Printf("事务执行出错: %v, 正在回滚事务\n", err)
		rollbackErr := u.Rollback()
		if rollbackErr != nil {
			fmt.Printf("回滚事务失败: %v\n", rollbackErr)
			return fmt.Errorf("原始错误: %v, 回滚失败: %w", err, rollbackErr)
		}
		return err
	}

	// fmt.Println("事务执行成功，正在提交")
	return u.Commit()
}

// UpdateFieldsById(current.Id, map[string]interface{}{
func (r *Repository[T]) UpdateFieldsById(id int64, fields map[string]interface{}) error {
	if r.uow != nil {
		return r.UpdateFieldsByIdWithTx(id, fields)
	}
	updateExp := make(map[string]interface{})

	for field, value := range fields {
		// 使用 exp.LiteralExpression 接口类型来判断
		if _, ok := value.(exp.LiteralExpression); ok {
			updateExp[field] = value
		} else {
			updateExp[field] = value
		}
	}

	query := r.dialect.Update(r.table).Set(updateExp).Where(goqu.Ex{"id": id})
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(sql, args...)
	return err
}

func (r *Repository[T]) UpdateFieldsByIds(ids []int64, fields map[string]interface{}) error {
	return r.UpdateFieldsByCondition(goqu.Ex{"id": ids}, fields)
}

// UpdateFieldsByIdWithTx
func (r *Repository[T]) UpdateFieldsByIdWithTx(id int64, fields map[string]interface{}) error {
	updateExp := make(map[string]interface{})
	for field, value := range fields {
		if _, ok := value.(exp.LiteralExpression); ok {
			updateExp[field] = value
		} else {
			updateExp[field] = value
		}
	}
	query := r.dialect.Update(r.table).Set(updateExp).Where(goqu.Ex{"id": id})
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	if r.uow != nil {
		fmt.Printf("事务更新SQL - 表:%s, SQL:%s, 参数:%v\n", r.table, sql, args)
		result, err := r.uow.GetTx().Exec(sql, args...)
		if err == nil {
			affected, _ := result.RowsAffected()
			fmt.Printf("事务更新结果 - 影响行数:%d\n", affected)
			if affected == 0 {
				fmt.Printf("警告: 更新操作未影响任何行\n")
			}
		} else {
			fmt.Printf("事务更新失败 - 错误:%v\n", err)
		}
		return err
	} else {
		_, err = r.db.Exec(sql, args...)
		return err
	}
}

// ScanTx(ctx context.Context, dest interface{}) error
func (r *Repository[T]) ScanTx(ctx context.Context, dest interface{}) error {
	query := r.dialect.From(r.table)
	sql, args, err := query.ToSQL()
	if err != nil {
		return err
	}
	return r.db.QueryRowxContext(ctx, sql, args...).StructScan(dest)
}

// ScanInt64Slice() ([]int64, error)
func (r *Repository[T]) ScanInt64Slice() ([]int64, error) {
	query := r.dialect.From(r.table)
	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []int64
	for rows.Next() {
		var val int64
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		result = append(result, val)
	}
	return result, nil
}

// ScanFloat64() (float64, error)
func (r *Repository[T]) ScanFloat64() (float64, error) {
	query := r.dialect.From(r.table)
	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, err
	}
	var result float64
	err = r.db.QueryRow(sql, args...).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// 写一个方法根据条件查询单个对象
func (r *Repository[T]) QuerySingle(condition goqu.Ex) (*T, error) {
	query := r.dialect.From(r.table).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = r.db.QueryRow(sql, args...).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// QuerySingleTx 带事务的 工作单元
func (r *Repository[T]) QuerySingleTx(ctx context.Context, condition goqu.Ex) (*T, error) {
	query := r.dialect.From(r.table).Where(condition)
	sql, args, err := query.ToSQL()
	if err != nil {
		return nil, err
	}
	var result T
	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func (r *Repository[T]) ToSQL() (sql string, params []interface{}, err error) {
	query := r.dialect.From(r.table)
	query1, args, err := query.ToSQL()
	return query1, args, err
}

func (r *Repository[T]) CreateAndReturnIDWithTx(entity *T) (int64, error) {
	// 构造插入语句
	query := r.dialect.Insert(r.table).Rows(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("生成插入SQL失败: %w", err)
	}

	// 通过事务执行插入
	result, err := r.uow.GetTx().Exec(sql, args...)
	if err != nil {
		return 0, fmt.Errorf("插入记录失败: %w", err)
	}

	// 获取自增ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取自增ID失败: %w", err)
	}

	return id, nil
}

func (r *Repository[T]) CreateAndReturnID(entity *T) (int64, error) {
	// 如果有工作单元，调用 CreateAndReturnIDWithTx
	if r.uow != nil {
		return r.CreateAndReturnIDWithTx(entity)
	}

	// 构造插入语句
	query := r.dialect.Insert(r.table).Rows(entity)
	sql, args, err := query.ToSQL()
	if err != nil {
		return 0, fmt.Errorf("生成插入SQL失败: %w", err)
	}

	// 执行插入并获取结果
	result, err := r.db.Exec(sql, args...)
	if err != nil {
		return 0, fmt.Errorf("插入记录失败: %w", err)
	}

	// 获取自增ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取自增ID失败: %w", err)
	}

	return id, nil
}
