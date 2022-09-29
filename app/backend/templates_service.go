package backend

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"os"
	"path/filepath"
)

type ITemplateService interface {
	Format(templateFileName string, arg interface{}) string
	Templates() map[string]string
}

type TemplateServiceImpl struct {
	ITemplateService

	Config    *core.Config
	templates map[string]string
}

func (c *TemplateServiceImpl) Templates() map[string]string {
	return c.templates
}

func (c *TemplateServiceImpl) Init() error {
	c.templates = map[string]string{}
	templatesDir := c.Config.GetOnBaseWorkDir("manual_email_templates")
	files, err := os.ReadDir(templatesDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		fBytes, err := os.ReadFile(filepath.Join(templatesDir, f.Name()))
		if err != nil {
			return err
		}
		c.templates[f.Name()] = string(fBytes)
	}
	return nil
}

func (c *TemplateServiceImpl) Format(template string, arg interface{}) string {
	template = c.templates[template]
	return utils.Format(template).Exec(arg)
}
