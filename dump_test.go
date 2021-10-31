package notionhacks_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/jhchabran/notionhacks"
	"github.com/jomei/notionapi"
)

func TestMarkdown(t *testing.T) {
	c := qt.New(t)
	c.Run("xt", func(t *qt.C) {
		tests := []struct {
			input   notionapi.RichText
			want    string
			wantErr bool
		}{
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: nil,
				},
				want:    "foobar",
				wantErr: false,
			},
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: &notionapi.Annotations{Bold: true},
				},
				want:    "*foobar*",
				wantErr: false,
			},
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: &notionapi.Annotations{Italic: true},
				},
				want:    "_foobar_",
				wantErr: false,
			},
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: &notionapi.Annotations{Italic: true, Bold: true},
				},
				want:    "_*foobar*_",
				wantErr: false,
			},
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: &notionapi.Annotations{Code: true},
				},
				want:    "`foobar`",
				wantErr: false,
			},
			{
				input: notionapi.RichText{
					PlainText:   "foobar",
					Annotations: &notionapi.Annotations{Bold: true},
					Href:        "http://foo.com",
				},
				want:    "(*foobar*)[http://foo.com]",
				wantErr: false,
			},
		}
		for _, test := range tests {
			var iw notionhacks.IndentWriter
			err := notionhacks.RichText(test.input).AppendMarkdown(&iw)
			if !test.wantErr {
				c.Assert(err, qt.IsNil)
			}
			c.Assert(iw.String(), qt.Equals, test.want)
		}
	})
}
