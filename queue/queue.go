package queue

import (
	"container/list"
	"sync"
)

type Queue struct {
	data  *list.List
	size  int
	mutex sync.Mutex
}

func New() *Queue {
	queue := new(Queue)
	queue.data = list.New()
	queue.size = 100
	return queue
}

func Exist(queue *Queue, url string) bool {
	for e := queue.data.Front(); e != nil; e = e.Next() {
		if e.Value == url {
			return true
		}
	}

	return false
}

func Add(queue *Queue, url string) {
	if queue.data.Len() >= queue.size {
		// Control the size of the Queue
		queue.mutex.Lock()
		queue.data.Remove(queue.data.Back())
		queue.mutex.Unlock()
	}

	queue.mutex.Lock()
	queue.data.PushFront(url)
	queue.mutex.Unlock()
}
