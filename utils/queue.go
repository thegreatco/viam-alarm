package utils

import "errors"

type Queue struct {
	values []interface{}
	head   int
	tail   int
	size   int
	count  int
}

func NewQueue(length int) (*Queue, error) {
	if length <= 0 {
		return nil, errors.New("length must be greater than 0")
	}
	return &Queue{
		values: make([]interface{}, length),
		head:   0,
		tail:   0,
		size:   length,
		count:  0,
	}, nil
}

func (q *Queue) IsFull() bool {
	return q.size == q.count
}

func (q *Queue) IsEmpty() bool {
	return q.count == 0
}

func (q *Queue) Push(value interface{}) error {
	if q.IsFull() {
		return errors.New("queue is full")
	}
	q.values[q.tail] = value
	q.tail = (q.tail + 1) % q.size
	q.count++
	return nil
}

func (q *Queue) ReadAll() []interface{} {
	result := make([]interface{}, q.count)
	for i := 0; i < q.count; i++ {
		result[i] = q.values[(q.head+i)%q.size]
	}
	return result
}
