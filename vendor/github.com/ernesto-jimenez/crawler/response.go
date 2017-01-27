package crawler

import (
	"golang.org/x/net/html"
	"io"
	"net/url"
	"strings"
)

// Response has the details from crawling a single URL
type Response struct {
	URL         string  `json:"url"`
	OriginalURL string  `json:"original_url,omitempty"`
	Links       []Link  `json:"links"`
	Assets      []Asset `json:"assets"`

	request *Request
}

// Link contains the informaiton from a single `a` tag
type Link struct {
	// URL contains the href attribute of the link. e.g: <a href="{href}">...</a>
	URL string `json:"url"`
}

// Asset represents linked assets such as link, script and img tags
type Asset struct {
	// Tag used to link the asset
	Tag string `json:"tag"`

	// URL of the asset
	URL string `json:"url"`

	// Rel contains the text of the rel attribute
	Rel string `json:"rel,omitempty"`

	// Type contains the text of the type attribute
	Type string `json:"type,omitempty"`
}

// ReadResponse extracts links and assets from the HTML read form the given io
// reader and fills it in the response
func ReadResponse(baseURL string, r io.Reader, res *Response) error {
	base, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	res.URL = base.String()
	res.Links = nil
	res.Assets = nil
	node, err := html.Parse(r)
	if err != nil {
		return err
	}
	var dfWalk func(*html.Node)
	dfWalk = func(n *html.Node) {
		if link := extractLink(base, n); link != nil {
			res.Links = append(res.Links, *link)
		} else if assets := extractAssets(base, n); len(assets) > 0 {
			res.Assets = append(res.Assets, assets...)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			dfWalk(c)
		}
	}
	dfWalk(node)
	return nil
}

func extractLink(base *url.URL, n *html.Node) *Link {
	if n.Type != html.ElementNode {
		return nil
	}
	if n.Data != "a" {
		return nil
	}
	var link Link
	for _, attr := range n.Attr {
		if attr.Key != "href" {
			continue
		}
		v, err := url.Parse(strings.TrimSpace(attr.Val))
		if err != nil {
			continue
		}
		if v.Scheme != "" && v.Scheme != "http" && v.Scheme != "https" {
			continue
		}
		v = base.ResolveReference(v)
		v.Fragment = ""
		link.URL = v.String()
	}
	if link.URL == "" {
		return nil
	}
	return &link
}

func extractAssets(base *url.URL, n *html.Node) []Asset {
	if n.Type != html.ElementNode {
		return nil
	}
	fn, ok := extractTag[n.Data]
	if !ok {
		return nil
	}
	return fn(base, n)
}

func extractSimpleAsset(base *url.URL, n *html.Node) []Asset {
	var asset Asset
	asset.Tag = n.Data
	var srcset string
	for _, attr := range n.Attr {
		switch attr.Key {
		case "src", "href":
			v, err := url.Parse(strings.TrimSpace(attr.Val))
			if err != nil {
				continue
			}
			asset.URL = base.ResolveReference(v).String()
		case "srcset":
			srcset = attr.Val
		case "rel":
			asset.Rel = attr.Val
		case "type":
			asset.Type = attr.Val
		}
	}
	if srcset == "" && asset.URL == "" {
		return nil
	}
	if srcset == "" {
		return []Asset{asset}
	}
	return assetSrcset(base, asset, srcset)
}

func assetSrcset(base *url.URL, asset Asset, srcset string) []Asset {
	srcs := strings.Split(srcset, ",")
	assets := make([]Asset, 0, len(srcs))
	for _, src := range srcs {
		src = strings.TrimSpace(src)
		parts := strings.Split(src, " ")
		u := parts[0]
		v, err := url.Parse(strings.TrimSpace(u))
		if err != nil {
			continue
		}
		asset.URL = base.ResolveReference(v).String()
		if asset.URL == "" {
			continue
		}
		assets = append(assets, asset)
	}
	return assets
}

func extractSourceAsset(base *url.URL, n *html.Node) []Asset {
	if n.Parent == nil {
		return nil
	}
	if n.Parent.Data != "picture" && n.Parent.Data != "video" {
		return nil
	}
	assets := extractSimpleAsset(base, n)
	for i := range assets {
		assets[i].Tag = n.Parent.Data + ">" + n.Data
	}
	return assets
}

func extractImgAsset(base *url.URL, n *html.Node) []Asset {
	if n.Parent != nil && n.Parent.Data == "picture" {
		return extractSourceAsset(base, n)
	}
	return extractSimpleAsset(base, n)
}

func extractLinkAsset(base *url.URL, n *html.Node) []Asset {
	assets := extractSimpleAsset(base, n)
	res := assets[0:0]
	for _, asset := range assets {
		if asset.Rel == "stylesheet" {
			res = append(res, asset)
		}
	}
	return res
}

type extractAssetFunc func(*url.URL, *html.Node) []Asset

var extractTag = map[string]extractAssetFunc{
	"link":   extractLinkAsset,
	"source": extractSourceAsset,
	"img":    extractImgAsset,
	"script": extractSimpleAsset,
	"video":  extractSimpleAsset,
}
