package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Simple is responsible of running a crawl, allowing you to queue new URLs to
// be crawled and build requests to be crawled.
type Simple struct {
	opts []Option
}

// New initialises a new crawl runner
func New(opts ...Option) (*Simple, error) {
	return &Simple{
		opts: opts,
	}, nil
}

// CrawlFunc is the type of the function called for each webpage visited by
// Crawl. The incoming url specifies which url was fetched, while res contains
// the response of the fetched URL if it was successful. If the fetch failed,
// the incoming error will specify the reason and res will be nil.
//
// Returning SkipURL will avoid queing up the resources links to be crawled.
//
// Returning any other error from the function will immediately stop the crawl.
type CrawlFunc func(url string, res *Response, err error) error

// SkipURL can be returned by CrawlFunc to avoid crawling the links from the given url
var SkipURL = errors.New("skip URL")

// Crawl will fetch all the linked websites starting from startURL and invoking
// crawlFn for each fetched url with either the response or the error.
//
// It will return an error if the crawl was prematurely stopped or could not be
// started.
//
// Crawl will always add WithOneRequestPerURL to the options of the worker to
// avoid infinite loops.
func (s *Simple) Crawl(startURL string, crawlFn CrawlFunc) error {
	req, err := NewRequest(startURL)
	if err != nil {
		return err
	}

	// initialise the queue
	queue := NewInMemoryQueue()
	queue.PushBack(req)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.opts = append(s.opts, WithOneRequestPerURL())

	w, err := NewWorker(crawlFn, s.opts...)
	if err != nil {
		return err
	}

	return w.Run(ctx, queue)
}

func fetch(ctx context.Context, c *http.Client, req *Request) (*Response, error) {
	url := req.URL.String()
	httpReq, err := http.NewRequest(http.MethodGet, req.URL.String(), nil)
	if err != nil {
		return nil, err
	}
	httpReq = httpReq.WithContext(ctx)

	httpRes, err := c.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()
	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s for %s", httpRes.Status, url)
	}
	res := Response{
		request: req,
	}
	finalURL := httpRes.Request.URL.String()
	err = ReadResponse(finalURL, httpRes.Body, &res)
	if finalURL != url {
		res.OriginalURL = url
	}
	return &res, nil
}

func nextRequest(res *Response, href string) (*Request, error) {
	req, err := NewRequest(href)
	if err != nil {
		return nil, err
	}
	req.depth = res.request.depth + 1
	return req, nil
}
