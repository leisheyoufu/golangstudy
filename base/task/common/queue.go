package common

import (
	"container/list"
)

type Queue struct {
	lst *list.List
}

func (q *Queue) Push(elem interface{}) {
	q.lst.PushBack(elem)
}

func (q *Queue) Pop() {
	elem := q.lst.Front()
	q.lst.Remove(elem)
}

func (q *Queue) Top() (elem interface{}) {
	return q.lst.Front()
}
