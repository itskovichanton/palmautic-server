package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
)

const (
	ErrReasonFeatureUnaccessable = "REASON_FEATURE_UNACCESSABLE"

	FeatureNameEmail     = "email"
	FeatureNameB2BSearch = "b2b-search"
)

type IFeatureAccessService interface {
	CheckFeatureAccessableEmail(accountId entities.ID) error
	NotifyFeatureUsedEmail(accountId entities.ID)
	CheckFeatureAccessableB2BSearch(accountId entities.ID) error
	NotifyFeatureUsedB2BSearch(accountId entities.ID)
}

type FeatureAccessServiceImpl struct {
	IFeatureAccessService

	UserRepo   IUserRepo
	TariffRepo ITariffRepo
	EventBus   EventBus.Bus
}

func (c *FeatureAccessServiceImpl) Init() {

}

func (c *FeatureAccessServiceImpl) CheckFeatureAccessableEmail(accountId entities.ID) error {
	return c.checkFeatureAccessable(accountId, func(account *entities.User) bool {
		return account.Tariff.FeatureAbilities.MaxEmailsPerDay > 0
	}, FeatureNameEmail, "Количество доступных Email иссякло. Оно восстанавливается каждые сутки.")
}

func (c *FeatureAccessServiceImpl) checkFeatureAccessable(accountId entities.ID, isFeatureAccessable func(account *entities.User) bool, featureName, errMsg string) error {
	account := c.UserRepo.FindById(accountId)
	if !isFeatureAccessable(account) {
		c.EventBus.Publish(FeatureUnaccessableByTariff, account, featureName)
		return errs.NewBaseErrorWithReason(errMsg, ErrReasonFeatureUnaccessable)
	}
	return nil
}

func (c *FeatureAccessServiceImpl) CheckFeatureAccessableB2BSearch(accountId entities.ID) error {
	return c.checkFeatureAccessable(accountId, func(account *entities.User) bool {
		return account.Tariff.FeatureAbilities.B2B && account.Tariff.FeatureAbilities.MaxB2BSearches > 0
	}, FeatureNameB2BSearch, "Количество поисков B2B иссякло. Оно восстанавливается каждые сутки.")
}

func (c *FeatureAccessServiceImpl) NotifyFeatureUsedB2BSearch(accountId entities.ID) {
	c.notifyFeatureUsed(accountId, func(account *entities.User) {
		if account.Tariff.FeatureAbilities.MaxB2BSearches >= 1 {
			account.Tariff.FeatureAbilities.MaxB2BSearches--
		}
	})
}

func (c *FeatureAccessServiceImpl) NotifyFeatureUsedEmail(accountId entities.ID) {
	c.notifyFeatureUsed(accountId, func(account *entities.User) {
		if account.Tariff.FeatureAbilities.MaxEmailsPerDay >= 1 {
			account.Tariff.FeatureAbilities.MaxEmailsPerDay--
		}
	})
}

func (c *FeatureAccessServiceImpl) notifyFeatureUsed(accountId entities.ID, afterFeatureUsed func(a *entities.User)) {
	account := c.UserRepo.FindById(accountId)
	if account != nil {
		afterFeatureUsed(account)
	}
}
