package crawler

import (
	"container/list"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Runner is responsible of running a crawl, allowing you to queue new URLs to
// be crawled and build requests to be crawled.
type Runner struct {
	client     *http.Client
	maxDepth   int
	checkFetch CheckFetchFunc
}

// New initialises a new crawl runner
func New(opts ...Option) (*Runner, error) {
	o := options{
		client: http.DefaultClient,
	}

	for _, opt := range opts {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	return &Runner{
		maxDepth:   o.maxDepth,
		client:     o.client,
		checkFetch: o.checkFetch,
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

// CheckFetchFunc is used to check whether a page should be fetched during the
// crawl or not
type CheckFetchFunc func(*url.URL) bool

// SkipURL can be returned by CrawlFunc to avoid crawling the links from the given url
var SkipURL = errors.New("skip URL")

// Crawl will fetch all the linked websites starting from startURL and invoking
// crawlFn for each fetched url with either the response or the error.
//
// It will return an error if the crawl was prematurely stopped or could not be
// started.
func (r *Runner) Crawl(startURL string, crawlFn CrawlFunc) error {
	req, err := NewRequest(startURL)
	if err != nil {
		return err
	}

	// initialise the queue
	queued := make(map[string]bool)
	queue := list.New()

	// add the first URL
	queue.PushBack(req)
	queued[req.URL.String()] = true

	// crawl
	for e := queue.Front(); e != nil; e = queue.Front() {
		queue.Remove(e)
		req := e.Value.(*Request)
		if r.checkFetch != nil && !r.checkFetch(req.URL) {
			continue
		}
		res, err := r.fetch(req)
		// call the CrawlFunc for each fetched url
		// note this err is scoped to the if and does not override the previous declaration
		if err := crawlFn(req.URL.String(), res, err); err == SkipURL {
			continue
		} else if err != nil {
			return err
		}
		// continue if there was an error crawlking
		if err != nil {
			continue
		}
		// Mark the response as visited since it might be different to the original URL due to redirects
		queued[res.URL] = true
		if r.maxDepth > 0 && r.maxDepth <= req.depth {
			continue
		}
		for _, link := range res.Links {
			if req, err := r.nextRequest(res, link.URL); err == nil && !queued[req.URL.String()] {
				queue.PushBack(req)
				queued[req.URL.String()] = true
			}
		}
	}
	return nil
}

func (r *Runner) fetch(req *Request) (*Response, error) {
	url := req.URL.String()
	httpRes, err := r.client.Get(url)
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

func (*Runner) nextRequest(res *Response, href string) (*Request, error) {
	req, err := NewRequest(href)
	if err != nil {
		return nil, err
	}
	req.depth = res.request.depth + 1
	return req, nil
}
