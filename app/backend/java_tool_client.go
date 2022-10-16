package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/echo-http"
	"net/http"
	"salespalm/server/app/entities"
	"strings"
)

type IJavaToolClient interface {
	FindEmail(params *FindEmailParams) ([]*FindEmailResult, error)
}

type JavaToolClientImpl struct {
	IJavaToolClient

	HttpClient *http.Client
	Config     *core.Config
	url        string
}

type FindEmailOrder struct {
	Subject, From []string
	MaxCount      int
}

type FindEmailResult struct {
	Subject, From string
	ContentParts  []*ContentPart
}

func (r FindEmailResult) DetectBounce() bool {

	if strings.Contains(strings.ToUpper(r.From), "DAEMON") {
		return true
	}

	for _, p := range r.ContentParts {
		if strings.Contains(p.Content, "не найден") || strings.Contains(p.Content, "не доставлено") {
			return true
		}
	}

	return false
}

type ContentPart struct {
	Content, ContentType string
}

type FindEmailParams struct {
	Access *entities.InMailSettings
	Order  *FindEmailOrder
}

func (c *JavaToolClientImpl) Init() {
	c.url = c.Config.GetStr("java-tools", "url")
}

func (c *JavaToolClientImpl) FindEmail(params *FindEmailParams) ([]*FindEmailResult, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
	println(string(b.Bytes()))
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%v/%v", c.url, "mail/scan"), b)
	request.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	if err != nil {
		return nil, err
	}
	resp, err := c.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}

	var r []*FindEmailResult
	_, err = frmclient.ExecuteWidthFrmAPI(resp, &r)
	return r, err
}
