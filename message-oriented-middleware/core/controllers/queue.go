package controllers

import (
	"fmt"
)

const (
	defaultQueueSize = 100
)

type Queue struct {
	data   chan string
	length int
}

func NewQueue() Queue {
	return Queue{
		data:   make(chan string, defaultQueueSize),
		length: defaultQueueSize,
	}
}

func (q *Queue) Add(item string) error {
	if len(q.data) >= q.length {
		return fmt.Errorf("queue has been full")
	}
	q.data <- item
	return nil
}

func (q *Queue) Get() string {
	return <-q.data
}
