package crawler_test

import (
	"fmt"

	"github.com/ernesto-jimenez/crawler"
)

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
