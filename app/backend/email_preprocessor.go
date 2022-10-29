package backend

import (
	"bytes"
	"fmt"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/itskovichanton/server/pkg/server"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"salespalm/server/app/entities"
	"strings"
)

type IEmailProcessorService interface {
	Process(result *FindEmailResult, accountId entities.ID)
}

type EmailProcessorServiceImpl struct {
	IEmailProcessorService
	Config *server.Config
}

func (c *EmailProcessorServiceImpl) replace(n *html.Node, accountId entities.ID) {
	if n.Type == html.ElementNode {
		if n.Data == "img" {
			srcAttrIndexToReplace := slices.IndexFunc(n.Attr, func(attr html.Attribute) bool { return attr.Key == "src" && !strings.HasPrefix(attr.Val, "http") })
			if srcAttrIndexToReplace > -1 {
				altAttrIndex := slices.IndexFunc(n.Attr, func(attr html.Attribute) bool { return attr.Key == "alt" && len(attr.Val) > 0 })
				if altAttrIndex >= 1 {
					altAttr := n.Attr[altAttrIndex]
					n.Attr[srcAttrIndexToReplace] = html.Attribute{Key: "src", Val: fmt.Sprintf("%v/getFile?key=%v__%v", c.Config.Server.GetUrl(), accountId, altAttr.Val)}
				}
			}
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		c.replace(child, accountId)
	}
}

func (c *EmailProcessorServiceImpl) Process(result *FindEmailResult, accountId entities.ID) {
	htmlPart := result.ContentParts[0].Content
	root, err := html.Parse(strings.NewReader(htmlPart))
	if err != nil {
		return
	}
	c.replace(root, accountId)
	buf := &bytes.Buffer{}
	if err = html.Render(buf, root); err == nil {
		result.ContentParts[0].Content = utils.HtmlUnescaper.Replace(buf.String())
	}
}
