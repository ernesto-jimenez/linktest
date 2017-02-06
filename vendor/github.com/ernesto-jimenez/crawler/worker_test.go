package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//go:generate goautomock Queue

func TestWorkerStopsOnContextCancelation(t *testing.T) {
	r := require.New(t)

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())

	q := NewQueueMock()
	q.PopFrontFunc = func() (*Request, error) {
		return NewRequest(s.URL)
	}
	q.PushBackFunc = func(*Request) error { return nil }

	w, err := NewWorker(func(url string, res *Response, err error) error {
		if q.PopFrontTotalCalls() == 2 {
			cancel()
		}
		return err
	})
	r.NoError(err)

	err = w.Run(ctx, q)
	r.Equal(context.Canceled, err)
	r.Equal(2, q.PopFrontTotalCalls())
}

func TestWorkerCancelsHTTPRequest(t *testing.T) {
	r := require.New(t)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		select {
		case <-ctx.Done():
			r.Equal(context.Canceled, ctx.Err())
		case <-time.After(time.Second):
			http.Error(w, "should have cancelled", http.StatusInternalServerError)
		}
	}))
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())

	q := NewQueueMock()
	q.PopFrontFunc = func() (*Request, error) {
		cancel()
		return NewRequest(s.URL)
	}
	q.PushBackFunc = func(*Request) error { return nil }

	w, err := NewWorker(func(url string, res *Response, err error) error {
		return err
	})
	r.NoError(err)

	err = w.Run(ctx, q)
	r.Equal(context.Canceled, err)
	r.Equal(1, q.PopFrontTotalCalls())
}
