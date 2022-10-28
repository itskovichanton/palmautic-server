package backend

import (
	"github.com/asaskevich/EventBus"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"salespalm/server/app/entities"
)

const (
	ErrReasonFeatureUnaccessable = "REASON_FEATURE_UNACCESSABLE"

	FeatureNameEmail = "email"
)

type IFeatureAccessService interface {
	CheckFeatureAccessableEmail(accountId entities.ID) error
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
	account := c.UserRepo.FindById(accountId)
	if account.Tariff.FeatureAbilities.MaxEmailsPerDay <= 0 {
		c.EventBus.Publish(FeatureUnaccessableByTariff, account, FeatureNameEmail)
		return errs.NewBaseErrorWithReason("Количество доступных Email иссякло. Оно восстанавливается каждые сутки.", ErrReasonFeatureUnaccessable)
	}
	return nil
}
