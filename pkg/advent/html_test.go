package advent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func Test_getH2NodeFromHTML(t *testing.T) {
	type args struct {
		doc *html.Node
	}
	tests := []struct {
		name      string
		args      args
		want      *html.Node
		assertion require.ErrorAssertionFunc
	}{
		{
			name: "empty document",
			args: args{
				doc: &html.Node{},
			},
			want:      nil,
			assertion: require.Error,
		},
		{
			name: "only h2 node",
			args: args{
				doc: &html.Node{
					Parent: nil,
					FirstChild: &html.Node{
						Parent:      nil,
						FirstChild:  nil,
						LastChild:   nil,
						PrevSibling: nil,
						NextSibling: nil,
						Type:        0x1,
						DataAtom:    0x0,
						Data:        "--- DAY 42: FAKE TITLE ---",
						Namespace:   "",
						Attr:        []html.Attribute{},
					},
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        0x3,
					DataAtom:    0x2de02,
					Data:        "h2",
					Namespace:   "",
					Attr:        []html.Attribute{},
				},
			},
			want: &html.Node{
				Parent: nil,
				FirstChild: &html.Node{
					Parent:      nil,
					FirstChild:  nil,
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        0x1,
					DataAtom:    0x0,
					Data:        "--- DAY 42: FAKE TITLE ---",
					Namespace:   "",
					Attr:        []html.Attribute{},
				},
				LastChild:   nil,
				PrevSibling: nil,
				NextSibling: nil,
				Type:        0x3,
				DataAtom:    0x2de02,
				Data:        "h2",
				Namespace:   "",
				Attr:        []html.Attribute{},
			},
			assertion: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getH2NodeFromHTML(tt.args.doc)

			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_renderNode(t *testing.T) {
	type args struct {
		n *html.Node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "h2 node with day/title",
			args: args{
				n: &html.Node{
					Parent: nil,
					FirstChild: &html.Node{
						Parent:      nil,
						FirstChild:  nil,
						LastChild:   nil,
						PrevSibling: nil,
						NextSibling: nil,
						Type:        0x1,
						DataAtom:    0x0,
						Data:        "--- DAY 42: FAKE TITLE ---",
						Namespace:   "",
						Attr:        []html.Attribute{},
					},
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        0x3,
					DataAtom:    0x2de02,
					Data:        "h2",
					Namespace:   "",
					Attr:        []html.Attribute{},
				},
			},
			want: "<h2>--- DAY 42: FAKE TITLE ---</h2>",
		},
		{
			name: "h2 node with empty text node",
			args: args{
				n: &html.Node{
					Parent: nil,
					FirstChild: &html.Node{
						Parent:      nil,
						FirstChild:  nil,
						LastChild:   nil,
						PrevSibling: nil,
						NextSibling: nil,
						Type:        0x1,
						DataAtom:    0x0,
						Data:        "",
						Namespace:   "",
						Attr:        []html.Attribute{},
					},
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        0x3,
					DataAtom:    0x2de02,
					Data:        "h2",
					Namespace:   "",
					Attr:        []html.Attribute{},
				},
			},
			want: "<h2></h2>",
		},
		{
			name: "only text node",
			args: args{
				n: &html.Node{
					Parent:      nil,
					FirstChild:  nil,
					LastChild:   nil,
					PrevSibling: nil,
					NextSibling: nil,
					Type:        0x1,
					DataAtom:    0x0,
					Data:        "--- DAY 42: FAKE TITLE ---",
					Namespace:   "",
					Attr:        []html.Attribute{},
				},
			},
			want: "--- DAY 42: FAKE TITLE ---",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, renderNode(tt.args.n))
		})
	}
}

func Test_renderNodeWithPanic(t *testing.T) {
	tests := []struct {
		name string
		node *html.Node
	}{
		{"empty node", &html.Node{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				renderNode(tt.node)
			})
		})
	}
}
