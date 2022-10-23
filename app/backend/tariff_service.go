package backend

import (
	"salespalm/server/app/entities"
	"time"
)

type ITariffRepo interface {
	All() []*entities.Tariff
	Commons() *TariffCommons
	FindById(id entities.ID) *entities.Tariff
}

const (
	TariffIDBasic        = -555
	TariffIDBasic2       = -558
	TariffIDProfessional = -556
	TariffIDEnterprise   = -557
)

type TariffCommons struct {
	Tariffs []*entities.Tariff
}

type TariffRepoImpl struct {
	ITariffRepo

	tariffs   []*entities.Tariff
	tariffMap map[entities.ID]*entities.Tariff
}

func (c *TariffRepoImpl) Init() {
	c.tariffs = []*entities.Tariff{
		{Price: 2000, Creds: entities.IDWithName{Name: "Basic", Id: TariffIDBasic2}, Due: 14 * time.Hour * 24, FeatureAbilities: &entities.FeatureAbilities{MaxSequences: 2, MaxEmailsPerDay: 200, B2B: false}},
		{Price: 0, Creds: entities.IDWithName{Name: "Basic", Id: TariffIDBasic}, Due: 14 * time.Hour * 24, FeatureAbilities: &entities.FeatureAbilities{MaxSequences: 2, MaxEmailsPerDay: 200, B2B: false}},
		{Price: 6600, Creds: entities.IDWithName{Name: "Professional", Id: TariffIDProfessional}, Due: 30 * time.Hour * 24, FeatureAbilities: &entities.FeatureAbilities{MaxSequences: 999, MaxEmailsPerDay: 10000, B2B: true}},
		{Price: -1, Creds: entities.IDWithName{Name: "Enterprise", Id: TariffIDEnterprise}, Due: 365 * time.Hour * 24, FeatureAbilities: &entities.FeatureAbilities{MaxSequences: 999, MaxEmailsPerDay: 10000000, B2B: true}},
	}
	c.tariffMap = map[entities.ID]*entities.Tariff{}
	for _, tariff := range c.tariffs {
		c.tariffMap[tariff.Creds.Id] = tariff
	}
}

func (c *TariffRepoImpl) FindById(id entities.ID) *entities.Tariff {
	return c.tariffMap[id]
}

func (c *TariffRepoImpl) All() []*entities.Tariff {
	return c.tariffs
}

func (c *TariffRepoImpl) Commons() *TariffCommons {
	return &TariffCommons{
		Tariffs: c.All(),
	}
}
