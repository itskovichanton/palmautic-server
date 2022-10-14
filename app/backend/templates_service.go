package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"salespalm/server/app/entities"
	"strings"
)

type ITemplateService interface {
	Format(template string, accountId entities.ID, args map[string]interface{}) string
	Templates(accountId entities.ID) TemplatesMap
	CreateOrUpdate(entity entities.IBaseEntity, body string, arg ...interface{}) string
	Clear(accountId entities.ID)
	Commons(accountId entities.ID) *TemplateCommons
}

type TemplatesMap map[string]string

type TemplateServiceImpl struct {
	ITemplateService

	Config                  *core.Config
	AccountService          IUserService
	templates               *cache.Cache
	TemplateCompilerService ITemplateCompilerService
	templatesDir            string
	optimize                bool
}

func (c *TemplateServiceImpl) Clear(accountId entities.ID) {
	c.templates.Delete(entities.IDStr(accountId))
}

func (c *TemplateServiceImpl) CreateOrUpdate(entity entities.IBaseEntity, body string, arg ...interface{}) string {

	prefix := strings.ToLower(reflect.ValueOf(entity).Type().Elem().Name())
	postfix := strings.Join(cast.ToStringSlice(arg), "_")

	templateFileName := ""
	if c.optimize {
		templateFileName = utils.MD5(body) + ".html"
	} else {
		templateFileName = fmt.Sprintf("%v__acc%v_id%v_%v.html", prefix, entity.GetAccountId(), entity.GetId(), postfix)
	}

	templateFullFileName := filepath.Join(c.calcTemplateForAccountDir(entity), templateFileName)
	err := os.WriteFile(templateFullFileName, []byte(body), 0755)
	if err != nil {
		templateFullFileName = ""
	}
	templateName := calcTemplateName(templateFileName)
	c.Templates(entity.GetAccountId())[templateName] = body
	return templateName

}

func calcTemplateName(templateFileName string) string {
	nameSplitterIndex := strings.Index(templateFileName, "__")
	if nameSplitterIndex > -1 {
		return templateFileName[:nameSplitterIndex]
	}
	_, templateFileName = filepath.Split(templateFileName)
	templateFileName = templateFileName[:len(templateFileName)-len(path.Ext(templateFileName))]
	return templateFileName
}

func (c *TemplateServiceImpl) Templates(accountId entities.ID) TemplatesMap {
	key := entities.IDStr(accountId)
	templatesMapI, _ := c.templates.Get(key)
	if templatesMapI == nil {
		templatesMap := TemplatesMap{}
		templatesMapI = templatesMap
		c.fillTemplatesMap(accountId, templatesMap)
		c.templates.Set(key, templatesMap, cache.NoExpiration)
	}
	return templatesMapI.(TemplatesMap)
}

func (c *TemplateServiceImpl) Init() {
	c.templates = cache.New(cache.NoExpiration, cache.NoExpiration)
	c.templatesDir = c.Config.GetDir("manual_email_templates")
	c.optimize = c.Config.GetBool("templates", "optimize")
}

func (c *TemplateServiceImpl) Format(template string, accountId entities.ID, args map[string]interface{}) string {
	if strings.HasPrefix(template, "template") {
		template = strings.Split(template, ":")[1]
		if len(template) > 0 {
			template = c.Templates(accountId)[template]
		}
	}
	return utils.Format(template).Exec(c.prepareArgs(accountId, args))
}

func (c *TemplateServiceImpl) calcTemplateForAccountDir(entity entities.IBaseEntity) string {
	r := c.calcTemplatesDir(entity.GetAccountId())
	os.MkdirAll(r, 0755)
	return r
}

func (c *TemplateServiceImpl) fillTemplatesMap(accountId entities.ID, templatesMap TemplatesMap) {
	filepath.Walk(c.calcTemplatesDir(accountId), func(path string, f fs.FileInfo, err error) error {

		if f.IsDir() || strings.Contains(f.Name(), "disabled)") {
			return nil
		}

		fBytes, _ := os.ReadFile(path)
		templatesMap[calcTemplateName(f.Name())] = entities.RemoveHtmlIndents(string(fBytes))

		return nil
	})
}

func (c *TemplateServiceImpl) calcTemplatesDir(accountId entities.ID) string {
	return filepath.Join(c.templatesDir, fmt.Sprintf("%v", accountId))
}

func (c *TemplateServiceImpl) prepareArgs(accountId entities.ID, args map[string]interface{}) interface{} {
	if args == nil {
		args = map[string]interface{}{}
	}
	args["Me"] = c.AccountService.Accounts()[accountId]
	return args
}

func (c *TemplateServiceImpl) Commons(accountId entities.ID) *TemplateCommons {
	return &TemplateCommons{
		Cache:       c.Templates(accountId),
		Compiler:    c.TemplateCompilerService.Commons(),
		Marketplace: c.Templates(0),
	}
}

type TemplateCommons struct {
	Cache       TemplatesMap
	Compiler    *TemplateCompilerCommons
	Marketplace TemplatesMap
}
