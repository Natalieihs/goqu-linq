// contrib/dao/base/enumerable.go
package core

// IEnumerable 内存操作接口
type IEnumerable[T any] interface {
	Where(predicate func(T) bool) IEnumerable[T]
	Select(selector func(T) interface{}) IEnumerable[interface{}]
	OrderBy(less func(T, T) bool) IEnumerable[T]
	Skip(count int) IEnumerable[T]
	Take(count int) IEnumerable[T]
	FirstOrDefault(predicate func(T) bool) *T
	LastOrDefault(predicate func(T) bool) *T
	ToList() []T
	Count() int
	Any(predicate func(T) bool) bool
	All(predicate func(T) bool) bool
	Sum(selector func(T) float64) float64
	Max(selector func(T) float64) float64
	Min(selector func(T) float64) float64
	Average(selector func(T) float64) float64
	GroupBy(keySelector func(T) interface{}) map[interface{}][]T
	Distinct(comparer func(T, T) bool) IEnumerable[T]
}

// Enumerable 实现
type Enumerable[T any] struct {
	data []T
}

func NewEnumerable[T any](data []T) IEnumerable[T] {
	return &Enumerable[T]{data: data}
}

func (e *Enumerable[T]) Where(predicate func(T) bool) IEnumerable[T] {
	var result []T
	for _, item := range e.data {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return NewEnumerable(result)
}

// Select 投影
func (e *Enumerable[T]) Select(selector func(T) interface{}) IEnumerable[interface{}] {
	var result []interface{}
	for _, item := range e.data {
		result = append(result, selector(item))
	}
	return NewEnumerable(result)
}

// OrderBy 排序
func (e *Enumerable[T]) OrderBy(less func(T, T) bool) IEnumerable[T] {
	result := make([]T, len(e.data))
	copy(result, e.data)
	quickSort(result, less)
	return NewEnumerable(result)
}

func quickSort[T any](data []T, less func(T, T) bool) {
	if len(data) <= 1 {
		return
	}
	mid := data[0]
	head, tail := 0, len(data)-1
	for i := 1; i <= tail; {
		if less(data[i], mid) {
			data[i], data[head] = data[head], data[i]
			head++
			i++
		} else {
			data[i], data[tail] = data[tail], data[i]
			tail--
		}
	}
	quickSort(data[:head], less)
	quickSort(data[head+1:], less)
}

func (e *Enumerable[T]) Skip(count int) IEnumerable[T] {
	if count >= len(e.data) {
		return NewEnumerable[T](nil)
	}
	return NewEnumerable(e.data[count:])
}

func (e *Enumerable[T]) Take(count int) IEnumerable[T] {
	if count >= len(e.data) {
		return e
	}
	return NewEnumerable[T](e.data[:count])
}

func (e *Enumerable[T]) FirstOrDefault(predicate func(T) bool) *T {
	for _, item := range e.data {
		if predicate(item) {
			return &item
		}
	}
	return nil
}

func (e *Enumerable[T]) LastOrDefault(predicate func(T) bool) *T {
	for i := len(e.data) - 1; i >= 0; i-- {
		if predicate(e.data[i]) {
			return &e.data[i]
		}
	}
	return nil
}

func (e *Enumerable[T]) ToList() []T {
	return e.data
}

func (e *Enumerable[T]) Count() int {
	return len(e.data)
}

func (e *Enumerable[T]) Any(predicate func(T) bool) bool {
	for _, item := range e.data {
		if predicate(item) {
			return true
		}
	}
	return false
}

func (e *Enumerable[T]) All(predicate func(T) bool) bool {
	for _, item := range e.data {
		if !predicate(item) {
			return false
		}
	}
	return true
}

func (e *Enumerable[T]) Sum(selector func(T) float64) float64 {
	var sum float64
	for _, item := range e.data {
		sum += selector(item)
	}
	return sum
}

func (e *Enumerable[T]) Max(selector func(T) float64) float64 {
	if len(e.data) == 0 {
		return 0
	}
	max := selector(e.data[0])
	for _, item := range e.data {
		value := selector(item)
		if value > max {
			max = value
		}
	}
	return max
}

func (e *Enumerable[T]) Min(selector func(T) float64) float64 {
	if len(e.data) == 0 {
		return 0
	}
	min := selector(e.data[0])
	for _, item := range e.data {
		value := selector(item)
		if value < min {
			min = value
		}
	}
	return min
}

func (e *Enumerable[T]) Average(selector func(T) float64) float64 {
	if len(e.data) == 0 {
		return 0
	}
	sum := e.Sum(selector)
	return sum / float64(len(e.data))
}

func (e *Enumerable[T]) GroupBy(keySelector func(T) interface{}) map[interface{}][]T {
	result := make(map[interface{}][]T)
	for _, item := range e.data {
		key := keySelector(item)
		result[key] = append(result[key], item)
	}
	return result
}

func (e *Enumerable[T]) Distinct(comparer func(T, T) bool) IEnumerable[T] {
	var result []T
	for _, item := range e.data {
		exists := false
		for _, r := range result {
			if comparer(item, r) {
				exists = true
				break
			}
		}
		if !exists {
			result = append(result, item)
		}
	}
	return NewEnumerable(result)
}
