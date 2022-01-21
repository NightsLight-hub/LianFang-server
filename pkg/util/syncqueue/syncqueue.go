/*
Package syncqueue
@Time : 2022/1/7 20:16
@Author : sunxy
@File : syncqueue
@description:
*/
package syncqueue

import (
	"container/list"
	"fmt"
	"sync"
)

type SyncQueue struct {
	Data *list.List
	lock *sync.Mutex
}

func NewQueue() *SyncQueue {
	return &SyncQueue{
		Data: list.New(),
		lock: new(sync.Mutex),
	}
}

func (q *SyncQueue) Push(v interface{}) {
	defer q.lock.Unlock()
	q.lock.Lock()
	q.Data.PushFront(v)
}

func (q *SyncQueue) Remove(e *list.Element) interface{} {
	defer q.lock.Unlock()
	q.lock.Lock()
	return q.Data.Remove(e)
}

func (q *SyncQueue) Pop() interface{} {
	defer q.lock.Unlock()
	q.lock.Lock()
	iter := q.Data.Back()
	v := iter.Value
	q.Data.Remove(iter)
	return v
}

func (q *SyncQueue) dump() {
	for iter := q.Data.Back(); iter != nil; iter = iter.Prev() {
		fmt.Println("item:", iter.Value)
	}
}
