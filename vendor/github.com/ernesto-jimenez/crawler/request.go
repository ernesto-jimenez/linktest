package crawler

import (
	"net/url"
)

// Request is used to fetch a page and informoation about its resources
type Request struct {
	URL *url.URL

	depth int
}

// NewRequest initialises a new crawling request to extract information from a single URL
func NewRequest(uri string) (*Request, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	u.Fragment = ""
	if u.Path == "" {
		u.Path = "/"
	}
	return &Request{
		URL: u,
	}, nil
}
