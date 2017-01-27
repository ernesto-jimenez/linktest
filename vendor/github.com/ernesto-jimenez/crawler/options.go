package crawler

import (
	"errors"
	"fmt"
	"net/http"
)

// Option is used to provide optional configuration to a crawler
type Option func(*options) error

type options struct {
	maxDepth   int
	client     *http.Client
	checkFetch CheckFetchFunc
}

// WithHTTPClient sets the optional http client
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) error {
		opts.client = client
		return nil
	}
}

// WithMaxDepth sets the max depth of the crawl
func WithMaxDepth(depth int) Option {
	return func(opts *options) error {
		if depth < 0 {
			return fmt.Errorf("depth must be greater or equal than zero. was: %d", depth)
		}
		opts.maxDepth = depth
		return nil
	}
}

// WithCheckFetch takes CheckFetchFunc that will be run before fetching each page to check whether it should be fetched or not
func WithCheckFetch(fn CheckFetchFunc) Option {
	return func(opts *options) error {
		if opts.checkFetch != nil {
			return errors.New("already set up CheckFetch function")
		}
		opts.checkFetch = fn
		return nil
	}
}
