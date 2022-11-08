package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/go-co-op/gocron"
	"github.com/itskovichanton/core/pkg/core"
	"github.com/jinzhu/copier"
	"salespalm/server/app/entities"
	"time"
)

type IAccountingService interface {
	Tariffs(accountId entities.ID) []*entities.Tariff
	AssignTariff(accountId, tariffId entities.ID)
}

type AccountingServiceImpl struct {
	IAccountingService

	UserRepo      IUserRepo
	TariffRepo    ITariffRepo
	EventBus      EventBus.Bus
	CronScheduler *gocron.Scheduler
	Config        *core.Config
}

func (c *AccountingServiceImpl) Init() {
	recoverTariffCron := c.Config.GetStr("accounting", "recovertariffcron")
	if len(recoverTariffCron) == 0 {
		recoverTariffCron = "0 0 * * *"
	}
	c.CronScheduler.Cron(recoverTariffCron).Do(c.recoverAllUsersTariff)
}

func (c *AccountingServiceImpl) Tariffs(accountId entities.ID) []*entities.Tariff {
	account := c.UserRepo.FindById(accountId)

	if account.Tariff == nil || account.Tariff.Creds.Id == TariffIDBasic && !account.Tariff.Expired() {
		return []*entities.Tariff{c.TariffRepo.FindById(TariffIDBasic), c.TariffRepo.FindById(TariffIDProfessional), c.TariffRepo.FindById(TariffIDEnterprise)}
	}

	return []*entities.Tariff{c.TariffRepo.FindById(TariffIDBasic2), c.TariffRepo.FindById(TariffIDProfessional), c.TariffRepo.FindById(TariffIDEnterprise)}
}

func (c *AccountingServiceImpl) AssignTariff(accountId, tariffId entities.ID) {

	account := c.UserRepo.FindById(accountId)
	tariff := c.TariffRepo.FindById(tariffId)

	if account.Tariff != nil && account.Tariff.Creds.Id == tariffId {
		// Этот тариф уже установлен
		return
	}

	if account.Tariff == nil {
		account.Tariff = &entities.Tariff{}
	}
	c.recoverTariff(account, tariff)
	account.Tariff.DueTime = time.Now().Add(tariff.Due)

}

func (c *AccountingServiceImpl) recoverTariff(account *entities.User, from *entities.Tariff) {
	if account.Tariff == nil {
		account.Tariff = &entities.Tariff{}
	}
	copier.Copy(&account.Tariff, &from)
	c.EventBus.Publish(TariffUpdatedEventTopic, account)
}

func (c *AccountingServiceImpl) recoverAllUsersTariff() {
	for _, account := range c.UserRepo.Accounts() {
		if account.Tariff != nil {
			tariff := c.TariffRepo.FindById(account.Tariff.Creds.Id)
			c.recoverTariff(account, tariff)
		}
	}
}
