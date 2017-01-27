package crawler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadResponse(t *testing.T) {
	tests := []struct {
		base    string
		file    string
		expects Response
	}{
		{
			base: "https://base.test/path/to/request",
			file: "testdata/index.html",
			expects: Response{
				URL: "https://base.test/path/to/request",
				Links: []Link{
					{URL: "http://example.localhost/absolute/url"},
					{URL: "https://base.test/absolute/path"},
					{URL: "https://base.test/path/relative/path"},
					{URL: "https://base.test/path/to/request"},
					{URL: "https://base.test/path/to/link-with-anchor/test"},
				},
				Assets: []Asset{
					{Tag: "script", URL: "https://base.test/path/to/example/javascript.js", Type: "text/javascript"},
					{Tag: "link", URL: "http://ok.test/style.css", Rel: "stylesheet", Type: "text/css"},
					{Tag: "img", URL: "https://base.test/path/to/logo.svg"},
					{Tag: "script", URL: "https://base.test/path/to/in-body.js", Type: "text/javascript"},
				},
			},
		},
		{
			base: "https://example",
			file: "testdata/assets.html",
			expects: Response{
				URL: "https://example",
				Assets: []Asset{
					{Tag: "script", URL: "https://example/example/javascript.js", Type: "text/javascript"},
					{Tag: "link", URL: "http://ok.test/style.css", Rel: "stylesheet", Type: "text/css"},
					{Tag: "img", URL: "https://example/logo.svg"},
					{Tag: "script", URL: "https://example/in-body.js", Type: "text/javascript"},
					{Tag: "picture>source", URL: "https://example/images/kitten-stretching.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-stretching@1.5x.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-stretching@2x.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-sitting.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-sitting@1.5x.png"},
					{Tag: "picture>img", URL: "https://example/images/kitten-curled@1.5x.png"},
					{Tag: "picture>img", URL: "https://example/images/kitten-curled@2x.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-stretching.png"},
					{Tag: "picture>source", URL: "https://example/images/kitten-sitting.png"},
					{Tag: "picture>img", URL: "https://example/images/kitten-curled.png"},
					{Tag: "picture>source", URL: "https://example/images/butterfly.webp", Type: "image/webp"},
					{Tag: "picture>img", URL: "https://example/images/butterfly.jpg"},
					{Tag: "video", URL: "https://example/video.webm"},
					{Tag: "video>source", URL: "https://example/devstories.webm", Type: `video/webm;codecs="vp8, vorbis"`},
					{Tag: "video>source", URL: "https://example/devstories.mp4", Type: `video/mp4;codecs="avc1.42E01E, mp4a.40.2"`},
					{Tag: "video>source", URL: "https://example/devstories.webm#t=10,20", Type: `video/webm;codecs="vp8, vorbis"`},
					{Tag: "video>source", URL: "https://example/devstories.webm", Type: `video/webm;codecs="vp8, vorbis"`},
					{Tag: "video>source", URL: "https://example/devstories.mp4", Type: `video/mp4;codecs="avc1.42E01E, mp4a.40.2"`},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			f, err := os.Open(test.file)
			require.NoError(t, err)

			defer f.Close()
			var actual Response
			err = ReadResponse(test.base, f, &actual)
			require.NoError(t, err)

			require.Equal(t, test.expects, actual)
		})
	}
}
