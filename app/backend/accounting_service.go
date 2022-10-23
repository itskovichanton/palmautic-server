package backend

import (
	"github.com/asaskevich/EventBus"
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

	UserRepo   IUserRepo
	TariffRepo ITariffRepo
	EventBus   EventBus.Bus
}

func (c *AccountingServiceImpl) Init() {

}

func (c *AccountingServiceImpl) Tariffs(accountId entities.ID) []*entities.Tariff {
	account := c.UserRepo.Accounts()[accountId]

	if account.Tariff == nil || account.Tariff.Creds.Id == TariffIDBasic && !account.Tariff.Expired() {
		return []*entities.Tariff{c.TariffRepo.FindById(TariffIDBasic), c.TariffRepo.FindById(TariffIDProfessional), c.TariffRepo.FindById(TariffIDEnterprise)}
	}

	return []*entities.Tariff{c.TariffRepo.FindById(TariffIDBasic2), c.TariffRepo.FindById(TariffIDProfessional), c.TariffRepo.FindById(TariffIDEnterprise)}
}

func (c *AccountingServiceImpl) AssignTariff(accountId, tariffId entities.ID) {

	account := c.UserRepo.Accounts()[accountId]
	tariff := c.TariffRepo.FindById(tariffId)

	if account.Tariff != nil && account.Tariff.Creds.Id == tariffId {
		// Этот тариф уже установлен
		return
	}

	if account.Tariff == nil {
		account.Tariff = &entities.Tariff{}
	}
	c.recoverTariff(account.Tariff, tariff)
	account.Tariff.DueTime = time.Now().Add(tariff.Due)

	c.EventBus.Publish(TariffUpdatedEventTopic, account)

}

func (c *AccountingServiceImpl) recoverTariff(to *entities.Tariff, from *entities.Tariff) {
	copier.Copy(&to, &from)
}
