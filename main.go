package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ernesto-jimenez/crawler"
	"github.com/ernesto-jimenez/httplogger"
)

var (
	path     string
	startURL string
	verbose  bool

	client  = &http.Client{}
	checked = map[string]error{}
)

func main() {
	var ()
	flag.BoolVar(&verbose, "verbose", false, "verbose logs information about every http request")
	flag.Parse()

	if verbose {
		client.Transport = httplogger.DefaultLoggedTransport
	}
	log.SetFlags(0)

	switch arg := flag.Arg(0); {
	case arg == "":
		path = "public"
	case strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://"):
		startURL = arg
		path = ""
	default:
		path = arg
	}

	if path != "" {
		_, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}

		s := httptest.NewServer(http.FileServer(http.Dir(path)))
		defer s.Close()
		startURL = s.URL
	}

	if _, err := url.Parse(startURL); err != nil {
		log.Fatal(err)
	}

	cr, err := crawler.New(
		crawler.WithHTTPClient(client),
		crawler.WithCheckFetch(func(u *url.URL) bool {
			return strings.HasPrefix(u.String(), startURL)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	var failures []failure

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		<-c
		cancel()
	}()

	err = cr.Crawl(startURL, func(url string, res *crawler.Response, err error) error {
		if err != nil {
			return err
		}
		if verbose {
			log.Printf("%s - Links: %d Assets: %d", url, len(res.Links), len(res.Assets))
		}
		in := strings.TrimPrefix(url, startURL)
		for _, u := range res.Links {
			err, ok := checked[u.URL]
			if !ok {
				err = checkURL(ctx, u.URL)
			}
			if err != nil {
				failures = append(failures, failure{
					inURL: in,
					err:   err,
					kind:  "link",
					res:   strings.TrimPrefix(u.URL, startURL),
				})
			}
			checked[u.URL] = err
		}
		for _, u := range res.Assets {
			err, ok := checked[u.URL]
			if !ok {
				err = checkURL(ctx, u.URL)
			}
			if err != nil {
				failures = append(failures, failure{
					inURL: in,
					err:   err,
					kind:  "resource",
					res:   strings.TrimPrefix(u.URL, startURL),
				})
			}
			checked[u.URL] = err
		}
		return ctx.Err()
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range failures {
		log.Println(f.Error())
	}

	if l := len(failures); l > 0 {
		log.Fatalf("%d broken", l)
	}
}

type failure struct {
	inURL string
	kind  string
	res   string
	err   error
}

func (f failure) Error() string {
	return fmt.Sprintf("in: %s %s: %s - %s", f.inURL, f.kind, f.res, f.err.Error())
}

func checkURL(ctx context.Context, u string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("responded with %s", res.Status)
	}
	return nil
}

func init() {
	// TODO: Remove this with Go 1.8
	// Disable http2 since the Go 1.7 client has an bug caulsing false positives
	os.Setenv("GODEBUG", "http2client=0")
}
