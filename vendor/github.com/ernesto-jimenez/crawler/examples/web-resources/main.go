package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ernesto-jimenez/crawler"
	"github.com/ernesto-jimenez/httplogger"
)

type result struct {
	Pages []*crawler.Response `json:"pages,omitempty"`
}

func main() {
	var (
		maxDepth     int
		includeHosts string
		excludeHosts string
		indentJSON   bool
		silent       bool
		outputFile   string
	)
	flag.IntVar(&maxDepth, "max-depth", 0, "max depth of links to follow with zero being unlimited (default is 0)")
	flag.StringVar(&includeHosts, "include-hosts", "", "list of hosts to crawl separated by commas (default is the host of the start URL)")
	flag.StringVar(&excludeHosts, "exclude-hosts", "", "list of hosts to skip in the crawl separated by commas")
	flag.BoolVar(&indentJSON, "indent", true, "whether to indent the produced JSON")
	flag.BoolVar(&silent, "silent", false, "whether to suppress progress output to STDERR")
	flag.StringVar(&outputFile, "output", "", "file to save the result of the crawl (default is STDOUT)")
	flag.Parse()

	var logOutput io.Writer
	if silent {
		logOutput = ioutil.Discard
	} else {
		logOutput = os.Stderr
	}
	log.SetOutput(logOutput)
	log.SetFlags(0)

	startURL := flag.Arg(0)
	if startURL == "" {
		log.Fatal("specify a start URL")
	}

	opts := []crawler.Option{
		crawler.WithHTTPClient(&http.Client{
			Transport: httplogger.DefaultLoggedTransport,
			Timeout:   5 * time.Second,
		}),
		crawler.WithMaxDepth(maxDepth),
		crawler.WithExcludedHosts(strings.Split(excludeHosts, ",")...),
	}

	if includeHosts != "" {
		opts = append(opts, crawler.WithAllowedHosts(strings.Split(includeHosts, ",")...))
	}

	cr, err := crawler.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	var output io.Writer
	if outputFile == "" {
		output = os.Stdout
	} else {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		output = f
	}

	var result result

	err = cr.Crawl(startURL, func(url string, res *crawler.Response, err error) error {
		if err != nil {
			log.Printf("error: %s", err.Error())
			return nil
		}
		result.Pages = append(result.Pages, res)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(output)
	if indentJSON {
		enc.SetIndent("", "\t")
	}
	err = enc.Encode(result)
	if err != nil {
		log.Fatal(err)
	}
}
