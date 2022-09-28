package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"os"
	"path/filepath"
)

type ITemplateService interface {
	GetTemplate(templateFileName string, arg interface{}) (string, error)
}

type TemplateServiceImpl struct {
	ITemplateService

	Config *core.Config
}

func (c *TemplateServiceImpl) GetTemplate(templateFileName string, arg interface{}) (string, error) {
	templateFileName = filepath.Join(c.Config.GetOnBaseWorkDir("manual_email_templates"), templateFileName)
	b, err := os.ReadFile(templateFileName)
	if err != nil {
		return "", nil
	}
	return utils.Format(string(b)).Exec(arg), nil
}
