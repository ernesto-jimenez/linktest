/*
* CODE GENERATED AUTOMATICALLY WITH github.com/ernesto-jimenez/goautomock
* THIS FILE MUST NEVER BE EDITED MANUALLY
 */

package crawler

// QueueMock mock
type QueueMock struct {
	calls map[string]int

	PopFrontFunc func() (*Request, error)

	PushBackFunc func(*Request) error
}

func NewQueueMock() *QueueMock {
	return &QueueMock{
		calls: make(map[string]int),
	}
}

// PopFront mocked method
func (m *QueueMock) PopFront() (*Request, error) {
	if m.PopFrontFunc == nil {
		panic("unexpected call to mocked method PopFront")
	}
	m.calls["PopFront"]++
	return m.PopFrontFunc()
}

// PopFrontCalls returns the amount of calls to the mocked method PopFront
func (m *QueueMock) PopFrontTotalCalls() int {
	return m.calls["PopFront"]
}

// PushBack mocked method
func (m *QueueMock) PushBack(p0 *Request) error {
	if m.PushBackFunc == nil {
		panic("unexpected call to mocked method PushBack")
	}
	m.calls["PushBack"]++
	return m.PushBackFunc(p0)
}

// PushBackCalls returns the amount of calls to the mocked method PushBack
func (m *QueueMock) PushBackTotalCalls() int {
	return m.calls["PushBack"]
}
