package crawler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryQueue(t *testing.T) {
	r := require.New(t)
	q := NewInMemoryQueue()

	r.Panics(func() {
		q.Init()
	})

	var (
		req *Request
		err error
	)

	req, err = q.PopFront()
	r.NoError(err)
	r.Nil(req)

	err = q.PushBack(&Request{depth: 1})
	r.NoError(err)
	err = q.PushBack(&Request{depth: 2})
	r.NoError(err)

	req, err = q.PopFront()
	r.NoError(err)
	r.Equal(1, req.depth)

	req, err = q.PopFront()
	r.NoError(err)
	r.Equal(2, req.depth)

	req, err = q.PopFront()
	r.NoError(err)
	r.Nil(req)
}
