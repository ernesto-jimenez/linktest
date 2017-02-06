package crawler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrawl(t *testing.T) {
	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	expected := []expectedPage{
		{url: s.URL + "/depth-one.html", totalLinks: 1},
		{url: s.URL + "/depth-two.html", totalLinks: 2},
		{url: s.URL + "/depth-three.html", totalLinks: 1},
		{url: s.URL + "/", totalLinks: 5, totalAssets: 4},
		{url: s.URL + "/depth-four.html", totalLinks: 1},
		{url: "http://example.localhost/absolute/url", hasError: true},
		{url: s.URL + "/absolute/path", hasError: true},
		{url: s.URL + "/relative/path", hasError: true},
		{url: s.URL + "/link-with-anchor/test", hasError: true},
	}

	testCrawl(t, s, "/depth-one.html", expected)
}

func TestMaxDepth(t *testing.T) {
	c, err := New(WithMaxDepth(1))
	require.NoError(t, err)

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	var n int

	err = c.Crawl(s.URL+"/depth-one.html", func(url string, res *Response, err error) error {
		n++
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, 2, n)
}

func TestCheckFetch(t *testing.T) {
	c, err := New(WithCheckFetch(func(req *Request) bool {
		return req.URL.Path == "/depth-one.html"
	}))
	require.NoError(t, err)

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	var n int

	err = c.Crawl(s.URL+"/depth-one.html", func(url string, res *Response, err error) error {
		n++
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, 1, n)
}

func TestAvoidVisitingTwice(t *testing.T) {
	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	expected := []expectedPage{
		{url: s.URL + "/start-cycle.html", totalLinks: 2},
		{url: s.URL + "/intermediate-cycle.html", totalLinks: 1},
		{url: s.URL + "/loop-cycle.html", totalLinks: 1},
	}

	testCrawl(t, s, "/start-cycle.html", expected)
}

func TestStopCrawl(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	var n int

	err = c.Crawl(s.URL+"/depth-one.html", func(url string, res *Response, err error) error {
		require.Equal(t, 0, n)
		n++
		return errors.New("failed")
	})
	require.Error(t, err)

	require.Equal(t, 1, n)
}

func TestCrawlInvalidStartHost(t *testing.T) {
	c, err := New()
	require.NoError(t, err)

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	err = c.Crawl(s.URL+"/non-existent", func(url string, res *Response, err error) error {
		require.Error(t, err)
		return nil
	})
	require.NoError(t, err)
}

func TestFetch(t *testing.T) {
	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	tests := []struct {
		path        string
		expectedURL string
	}{
		{path: "", expectedURL: s.URL + "/"},
		{path: "/", expectedURL: s.URL + "/"},
		{path: "/index.html", expectedURL: s.URL + "/"}, // FileServer will redirect to /
		{path: "/#with-fragment", expectedURL: s.URL + "/"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			req, err := NewRequest(s.URL + test.path)
			require.NoError(t, err)

			res, err := fetch(context.Background(), http.DefaultClient, req)
			require.NoError(t, err)

			require.Equal(t, test.expectedURL, res.URL)
		})
	}
}

type expectedPage struct {
	url         string
	totalLinks  int
	totalAssets int
	hasError    bool
}

func testCrawl(t *testing.T, s *httptest.Server, startPath string, expected []expectedPage) {
	c, err := New()
	require.NoError(t, err)

	var actual []expectedPage

	err = c.Crawl(s.URL+startPath, func(url string, res *Response, err error) error {
		result := expectedPage{url: url}
		if err == nil {
			result.hasError = false
			result.totalLinks = len(res.Links)
			result.totalAssets = len(res.Assets)
		} else {
			result.hasError = true
		}
		actual = append(actual, result)
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, expected, actual)
}
