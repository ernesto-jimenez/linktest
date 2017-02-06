package crawler

import (
	"context"
	"net/http"
)

// Runner defines the interface requred to run a crawl
type Runner interface {
	Run(context.Context, Queue) error
}

// Worker is used to run a crawl on a single goroutine
type Worker struct {
	client     *http.Client
	fn         CrawlFunc
	checkFetch CheckFetchStack
	maxDepth   int
}

// NewWorker initialises a goroutine
func NewWorker(fn CrawlFunc, opts ...Option) (*Worker, error) {
	o := options{
		client: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	return &Worker{
		client:     o.client,
		maxDepth:   o.maxDepth,
		checkFetch: CheckFetchStack(o.checkFetch),
		fn:         fn,
	}, nil
}

// Run starts processing requests from the queue
func (w *Worker) Run(ctx context.Context, q Queue) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		req, err := q.PopFront()
		if err != nil {
			return err
		}
		if req == nil {
			return nil
		}
		if !w.checkFetch.CheckFetch(req) {
			continue
		}
		res, err := fetch(ctx, w.client, req)
		if err := ctx.Err(); err != nil {
			return err
		}
		// call the CrawlFunc for each fetched url
		// note this err is scoped to the if and does not override the previous declaration
		if err := w.fn(req.URL.String(), res, err); err == SkipURL {
			continue
		} else if err != nil {
			return err
		}
		// continue if there was an error crawlking
		if err != nil {
			continue
		}
		// Mark the response as visited since it might be different to the original URL due to redirects
		if w.maxDepth > 0 && w.maxDepth <= req.depth {
			continue
		}
		for _, link := range res.Links {
			if req, err := nextRequest(res, link.URL); err == nil {
				q.PushBack(req)
			}
		}
	}
}
