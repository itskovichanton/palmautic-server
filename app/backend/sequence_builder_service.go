package backend

import (
	"fmt"
	"salespalm/server/app/entities"
)

type ISequenceBuilderService interface {
	Rebuild(sequence *entities.Sequence) (TemplatesMap, error)
	Log(spec *entities.SequenceSpec) (string, error)
}

type SequenceBuilderServiceImpl struct {
	ISequenceBuilderService

	TemplateService ITemplateService
}

func (c *SequenceBuilderServiceImpl) Rebuild(sequence *entities.Sequence) (TemplatesMap, error) {
	return nil, nil
}

func (c *SequenceBuilderServiceImpl) Log(spec *entities.SequenceSpec) (string, error) {
	return fmt.Sprintf(`<h3>FirstName: %v</h3>`, spec.Name), nil
}
