package cvvapi

import (
	"bytes"
	"io"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type htmlNode html.Node

func parseHtmlReader(reader io.Reader) *htmlNode {
	var node, err = html.Parse(reader)
	if err != nil {
		return &htmlNode{}
	}

	return (*htmlNode)(node)
}

func parseHtmlFragment(htmlStr string) *htmlNode {
	return parseHtmlReader(strings.NewReader(htmlStr))
}

func (o *htmlNode) IsEmpty() bool {
	if o == nil {
		return true
	}

	var buffer bytes.Buffer

	_ = html.Render(&buffer, (*html.Node)(o))

	return buffer.String() == ""
}

func (o *htmlNode) querySelector(query string) *htmlNode {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return &htmlNode{}
	}
	return (*htmlNode)(cascadia.Query((*html.Node)(o), sel))
}

func (o *htmlNode) querySelectorAll(query string) []*htmlNode {
	sel, err := cascadia.Parse(query)
	if err != nil {
		return []*htmlNode{}
	}
	var nodes = cascadia.QueryAll((*html.Node)(o), sel)
	var result = make([]*htmlNode, len(nodes))
	for idx, item:=range nodes {
		result[idx] = (*htmlNode)(item)
	}
	return result
}

func (o *htmlNode) getAttr(attrName string) string {
	if o == nil {
		return ""
	}

	for _, a := range o.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

func getHtmlNodeText(buffer *bytes.Buffer, node *html.Node) {
	if node.Type == html.TextNode {
		buffer.WriteString(node.Data)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		getHtmlNodeText(buffer, c)
	}
}

func (o *htmlNode) getText() string {
	if o == nil {
		return ""
	}
	
	var buffer bytes.Buffer

	getHtmlNodeText(&buffer, (*html.Node)(o))

	return buffer.String()
}

func (o *htmlNode) getHtml() string {
	if o == nil {
		return ""
	}
	
	var buffer bytes.Buffer

	var err = html.Render(&buffer, (*html.Node)(o))

	if err != nil {
		return ""
	}

	return buffer.String()
}