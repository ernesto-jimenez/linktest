package crawler

import (
	"container/list"
	"sync"
)

// Queue is used by workers to keep track of the urls that need to be fetched.
// Queue must be safe to use concurrently.
type Queue interface {
	PushBack(*Request) error
	PopFront() (*Request, error)
}

// InMemoryQueue holds a queue of items to be crawled in memory
type InMemoryQueue struct {
	mut   sync.Mutex
	queue *list.List
}

// NewInMemoryQueue returns an in memory queue ready to be used by different workers
func NewInMemoryQueue() *InMemoryQueue {
	q := &InMemoryQueue{}
	q.Init()
	return q
}

// Init is used to initialise the unexported fields on the InMemoryQueue. It is already called by NewInMemoryQueue and it only has to be called manually once if when initialising an InMemoryQueue with a literal.
//
// It will panic if called twice on the same queue.
func (q *InMemoryQueue) Init() {
	q.mut.Lock()
	defer q.mut.Unlock()

	if q.queue != nil {
		panic("init called on already initialised queue")
	}
	q.queue = list.New()
}

// PushBack adds a request to the queue
func (q *InMemoryQueue) PushBack(req *Request) error {
	q.queue.PushBack(req)
	return nil
}

// PopFront gets the next request from the queue.
// It will return a nil request and a nil error if the queue is empty.
func (q *InMemoryQueue) PopFront() (*Request, error) {
	v := q.queue.Front()
	if v == nil {
		return nil, nil
	}
	q.queue.Remove(v)
	return v.Value.(*Request), nil
}
