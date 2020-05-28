package controllers

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

const (
	defaultQueueSize = 100
)

type MsgQueue interface {
	Get() string
	Put(msg string) error
	Done()
	Forget()
}

type Queue struct {
	sync.Mutex
	list.List
	cursor *list.Element
	cap    int
}

func NewQueue(cap int) *Queue {
	return &Queue{
		cap: cap,
	}
}

func (q *Queue) Put(msg string) error {
	q.Lock()
	if q.List.Len() >= q.cap {
		q.Unlock()
		return fmt.Errorf("queue has been full")
	}
	q.List.PushBack(msg)
	q.Unlock()
	return nil
}

func (q *Queue) Forget() {
	q.Lock()
	if q.cursor != nil {
		q.List.Remove(q.cursor)
		q.cursor = nil
	}
	q.Unlock()
}

func (q *Queue) Done() {
	q.Lock()
	q.cursor = nil
	q.Unlock()
}

func (q *Queue) Get() string {
	for {
		q.Lock()
		if q.cursor != nil {
			q.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}
		e := q.List.Front()
		q.cursor = e
		q.Unlock()
		if e == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		return e.Value.(string)
	}
}

//type Queue struct {
//	data   chan string
//	length int
//}

//func NewQueue() Queue {
//	return Queue{
//		data:   make(chan string, defaultQueueSize),
//		length: defaultQueueSize,
//	}
//}
//
//func (q *Queue) Add(item string) error {
//	if len(q.data) >= q.length {
//		return fmt.Errorf("queue has been full")
//	}
//	q.data <- item
//	return nil
//}
//
//func (q *Queue) Get() string {
//	return <-q.data
//}
