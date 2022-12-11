package backend

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"salespalm/server/app/entities"
	"strings"
	"time"
)

type ISequenceBuilderService interface {
	Rebuild(sequence *entities.Sequence) (TemplatesMap, error)
	Log(spec *entities.SequenceSpec) (string, error)
}

type SequenceBuilderServiceImpl struct {
	ISequenceBuilderService

	TemplateService ITemplateService
	EventBus        EventBus.Bus
}

func (c *SequenceBuilderServiceImpl) Rebuild(r *entities.Sequence) (TemplatesMap, error) {
	m := r.Spec
	m.BaseEntity = r.BaseEntity
	r.Name = m.Name
	r.FolderID = m.FolderID
	r.Description = m.Description
	r.Model = &entities.SequenceModel{Steps: m.Model.Steps}
	usedTemplates := TemplatesMap{}
	for stepIndex, step := range r.Model.Steps {

		step.AccountId = r.AccountId

		// Устанавливаем время тасков
		if stepIndex > 0 {
			step.StartTime = r.Model.Steps[stepIndex-1].DueTime
		} else {
			// у 1го шага добавляем delay
			step.StartTime = step.StartTime.Add(time.Duration(step.Delay) * time.Second)
		}
		step.DueTime = step.StartTime.Add(5 * time.Minute)

		// Обновляем шаблоны email-тасков
		if step.HasTypeEmail() {
			if !strings.HasPrefix(step.Body, "template") {
				// сохраняем шаблон в папку
				templateName, err := c.TemplateService.CreateOrUpdate(step, step.Body, fmt.Sprintf("step%v", stepIndex))
				if err != nil {
					return usedTemplates, err
				}
				usedTemplates[templateName] = step.Body
				if len(templateName) > 0 {
					step.Body = "template:" + templateName
				}
			}
		}
	}

	c.EventBus.Publish(SequenceUpdatedEventTopic, r)

	return usedTemplates, nil
}

func (c *SequenceBuilderServiceImpl) Log(spec *entities.SequenceSpec) (string, error) {
	return fmt.Sprintf(`<h3>FirstName: %v</h3>`, spec.Name), nil
}
