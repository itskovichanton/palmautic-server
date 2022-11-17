package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/frmclient"
	"github.com/itskovichanton/echo-http"
	"net/http"
	"salespalm/server/app/entities"
	"strings"
)

type IJavaToolClient interface {
	FindEmail(params *FindEmailParams) (map[string]FindEmailResults, error)
	CheckAccess(access *EmailAccess) error
}

type JavaToolClientImpl struct {
	IJavaToolClient

	HttpClient *http.Client
	Config     *core.Config
	url        string
}

type EmailAccess struct {
	Login, Password, Server string
	Port                    int
}

func NewEmailAccessFromInMailSettings(emailSettings *entities.InMailSettings) *EmailAccess {
	return &EmailAccess{
		Login:    emailSettings.Login,
		Password: emailSettings.Password,
		Server:   emailSettings.ImapHost,
		Port:     emailSettings.ImapPort,
	}
}

type FindEmailOrder struct {
	Subjects, From []string
	MaxCount       int
	Instant        bool
}

type FindEmailResult struct {
	Subject, From string
	ContentParts  []*ContentPart
}

func (r *FindEmailResult) DetectBounce() bool {

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

func (r *FindEmailResult) PlainContent() string {
	if len(r.ContentParts) == 0 {
		return ""
	}
	part0 := r.ContentParts[0]
	if len(part0.PlainContent) > 0 {
		return part0.PlainContent
	}
	return strip.StripTags(part0.Content)
}

type ContentPart struct {
	Content, ContentType, FileName, PlainContent string
}

type FindEmailParams struct {
	Access *EmailAccess
	Orders map[string]*FindEmailOrder
}

func (c *JavaToolClientImpl) Init() {
	c.url = c.Config.GetStr("java-tools", "url")
}

type FindEmailResults []*FindEmailResult

func (r FindEmailResults) DetectBounce() bool {
	for _, result := range r {
		if result.DetectBounce() {
			return true
		}
	}
	return false
}

func (c *JavaToolClientImpl) FindEmail(params *FindEmailParams) (map[string]FindEmailResults, error) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(params)
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

	var r map[string]FindEmailResults
	if params.Orders == nil || len(params.Orders) == 0 {
		return r, nil
	}
	_, err = frmclient.ExecuteWidthFrmAPI(resp, &r)
	return r, err
}

func (c *JavaToolClientImpl) CheckAccess(access *EmailAccess) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(access)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", fmt.Sprintf("%v/%v", c.url, "mail/check"), b)
	request.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	if err != nil {
		return err
	}
	resp, err := c.HttpClient.Do(request)
	if err != nil {
		return err
	}
	_, err = frmclient.ExecuteWidthFrmAPI(resp, nil)
	return err
}
