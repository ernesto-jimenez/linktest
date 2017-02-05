# linktest

linktest is a tool to check all the links and resources a site points to are
working. It helps detect broken links or incorrect src urls.

The program will crawl all the local webpages accessible and check that all
links and resources return 200.

It can be used to check a static site, in which case linktest will start a
local server and crawl the site starting at `/`. This characteristic means it
can be used in your continuous integration server to check a `jekyll` or `hugo`
site has zero broken links and resources.


# Installation:

If you have never worked with golang:

```shell
# make sure Go and make are installed in your system
$ git clone github.com/ernesto-jimenez/linktest
$ cd linktest
$ make
# binary will be in ./bin/linktest
```

If you are a golang developer:

```shell
$ go get github.com/ernesto-jimenez/linktest
```

## Usage

```shell
# check a website
$ linktest http://gopheracademy.com
in: / link: https://gopheracademy.com/index.html - responded with 404 Not Found
in: / resource: https://gopheracademy.com/assets/img/ajax-loader.gif - responded with 404 Not Found
2 broken

# check a static site hosted on your filesystem
$ linktest dir

# check a static site on `./public`
$ linktest
```

