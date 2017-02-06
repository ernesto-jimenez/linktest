# crawler

[![Build Status](https://travis-ci.org/ernesto-jimenez/crawler.svg?branch=master)](https://travis-ci.org/ernesto-jimenez/crawler)
[![Coverage Status](https://coveralls.io/repos/ernesto-jimenez/crawler/badge.svg?branch=master&service=github)](https://coveralls.io/github/ernesto-jimenez/crawler?branch=master)
[![Go Report Card](https://goreportcard.com/badge/ernesto-jimenez/crawler)](https://goreportcard.com/ernesto-jimenez/crawler)

A simple package to quickly build programs that require crawling websites.

```
go get github.com/ernesto-jimenez/crawler
```


## Usage

[embedmd]:# (example_crawler_test.go /func Example/ $)
```go
func Example() {
	startURL := "https://godoc.org"

	cr, err := crawler.New()
	if err != nil {
		panic(err)
	}

	err = cr.Crawl(startURL, func(url string, res *crawler.Response, err error) error {
		if err != nil {
			fmt.Printf("error: %s", err.Error())
			return nil
		}
		fmt.Printf("%s - Links: %d Assets: %d\n", url, len(res.Links), len(res.Assets))
		return crawler.SkipURL
	})
	if err != nil {
		panic(err)
	}
	// Output:
	// https://godoc.org/ - Links: 39 Assets: 5
}
```
