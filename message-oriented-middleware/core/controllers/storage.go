package controllers

import "sync"

type QueueMap struct {
	keyMQueue map[string]Queue
	mutex     sync.RWMutex
}

func NewQueueMap() *QueueMap {
	return &QueueMap{
		keyMQueue: make(map[string]Queue),
	}
}

func (qm *QueueMap) Add(key string, queue Queue) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()
	qm.keyMQueue[key] = queue
}

func (qm *QueueMap) Get(key string) (Queue, bool) {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()
	queue, ok := qm.keyMQueue[key]
	return queue, ok
}
