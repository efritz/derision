package request

import "sync"

type (
	Log interface {
		Chan() <-chan *Request
		Copy(clear bool) []*Request
		Add(request *Request)
		Clear()
	}

	log struct {
		capacity     int
		requestChan  chan *Request
		requestSlice []*Request
		mutex        sync.RWMutex
	}
)

func NewLog(capacity int) *log {
	return &log{
		capacity:     capacity,
		requestChan:  make(chan *Request),
		requestSlice: []*Request{},
	}
}

func (l *log) Chan() <-chan *Request {
	return l.requestChan
}

func (l *log) Copy(clear bool) []*Request {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	requests := []*Request{}
	for _, request := range l.requestSlice {
		requests = append(requests, request)
	}

	if clear {
		l.clear()
	}

	return requests
}

func (l *log) Add(request *Request) {
	l.mutex.Lock()
	l.requestChan <- request
	l.requestSlice = append(l.requestSlice, request)
	l.prune()
	l.mutex.Unlock()
}

func (l *log) prune() {
	if l.capacity == 0 {
		return
	}

	for len(l.requestSlice) > l.capacity {
		l.requestSlice = l.requestSlice[1:]
	}
}

func (l *log) Clear() {
	l.mutex.Lock()
	l.clear()
	l.mutex.Unlock()
}

func (l *log) clear() {
	l.requestSlice = l.requestSlice[:0]
}
