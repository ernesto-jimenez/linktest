package crawler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Option is used to provide optional configuration to a crawler
type Option func(*options) error

type options struct {
	maxDepth   int
	client     *http.Client
	checkFetch []CheckFetchFunc
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
		opts.checkFetch = append(opts.checkFetch, fn)
		return nil
	}
}

// WithOneRequestPerURL adds a check to only allow URLs once
func WithOneRequestPerURL() Option {
	var mut sync.Mutex
	v := make(map[string]struct{})
	return WithCheckFetch(func(req *Request) bool {
		mut.Lock()
		defer mut.Unlock()
		_, ok := v[req.URL.String()]
		if ok {
			return false
		}
		v[req.URL.String()] = struct{}{}
		return true
	})
}

// WithAllowedHosts adds a check to only allow URLs with the given hosts
func WithAllowedHosts(hosts ...string) Option {
	m := make(map[string]struct{})
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		m[h] = struct{}{}
	}
	return WithCheckFetch(func(req *Request) bool {
		_, ok := m[req.URL.Host]
		return ok
	})
}

// WithExcludedHosts adds a check to only allow URLs with hosts other than the given ones
func WithExcludedHosts(hosts ...string) Option {
	m := make(map[string]struct{})
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		m[h] = struct{}{}
	}
	return WithCheckFetch(func(req *Request) bool {
		_, ok := m[req.URL.Host]
		return !ok
	})
}
