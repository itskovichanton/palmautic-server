package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"math/rand"
	"salespalm/server/app/entities"
	utils2 "salespalm/server/app/utils"
	"time"
)

type ITaskDemoService interface {
	GenerateTasks(count int, task *entities.Task) int
	FindRandomContact(accountId entities.ID) *entities.Contact
}

type TaskDemoServiceImpl struct {
	ITaskDemoService

	ContactService  IContactService
	TaskService     ITaskService
	SequenceService ISequenceService
	TemplateService ITemplateService
	AccountService  IUserService
	Config          *core.Config
	templates       map[string]string
}

func (c *TaskDemoServiceImpl) Init() error {
	rand.Seed(42)
	c.templates = c.TemplateService.Templates()
	return nil
}

func (c *TaskDemoServiceImpl) FindRandomContact(accountId entities.ID) *entities.Contact {
	return c.ContactService.GetByIndex(accountId, rand.Intn(1000))
}

func (c *TaskDemoServiceImpl) GenerateTasks(count int, task *entities.Task) int {

	generated := 0
	for i := 0; i < count; i++ {
		contact := c.FindRandomContact(task.AccountId)
		_, err := c.TaskService.CreateOrUpdate(c.generateRandomTask(contact, task))
		if err == nil {
			generated++
		}
	}
	return generated
}

func (c *TaskDemoServiceImpl) generateRandomTask(contact *entities.Contact, spec *entities.Task) *entities.Task {

	r := &entities.Task{
		BaseEntity: entities.BaseEntity{
			AccountId: spec.AccountId,
		},
		Sequence: &entities.IDAndTitle{
			ID: c.SequenceService.GetByIndex(spec.AccountId, rand.Intn(30)).GetId(),
		},
		Contact: contact,
	}

	types := c.TaskService.Commons(spec.AccountId).Types
	var taskType *entities.TaskType
	if len(spec.Type) == 0 {
		taskType = *utils2.RandomEntry(types)
	} else {
		taskType = types[spec.Type]
	}
	r.Type = taskType.Creds.Name
	if taskType.IsMessenger() && len(contact.Phone) == 0 {
		if rand.Intn(10) > 5 {
			taskType = entities.TaskTypeManualEmail
		} else {
			taskType = entities.TaskTypeLinkedin
		}
	}
	r.Type = taskType.Creds.Name
	r.Action = taskType.Actions[rand.Intn(len(taskType.Actions))].Name

	switch r.Type {
	case entities.TaskTypeWhatsapp.Creds.Name:
		r.Name = "Написать в Whatsapp"
		r.Description = "Написать в личное сообщение Whatsapp: " + utils2.FormatUrl("https://wa.me", contact.Phone)
		break
	case entities.TaskTypeTelegram.Creds.Name:
		r.Name = "Написать в Telegram"
		r.Description = "Написать в личное сообщение Telegram: " + utils2.FormatUrl("https://t.me", contact.Phone)
		break
	case entities.TaskTypeCall.Creds.Name:
		r.Name = "Позвонить по телефону"
		r.Description = "Позвонить по номеру телефона " + contact.Phone + ". Настройся на продуктивный лад 😊"
		break
	case entities.TaskTypeLinkedin.Creds.Name:
		c.updateLinkedinTask(r)
		break
	case entities.TaskTypeManualEmail.Creds.Name:
		r.Name = "Отправить письмо"
		r.Description = fmt.Sprintf("Отправить письмо для %v на %v", contact.Name, contact.Email)
		break
	}

	switch r.Type {
	case entities.TaskTypeManualEmail.Creds.Name:
		r.DueTime = time.Now().Add(20 * time.Minute)
		templateName := *utils2.RandomEntry(c.templates)
		r.Body = "template:" + templateName
		r.Subject = "Компания ITBest приглашает Вас на собеседование!"
		break
	case entities.TaskTypeCall.Creds.Name:
		r.Body = fmt.Sprintf(`Добрый день, я говорю с %v?. Отлично! Меня зовут Антон, я - менеджер по набору персонала в компании ITBestTech. Мы хотели бы пригласить Вас на собеседование. Как я понимаю, сейчас Вы трудоустроены в компании "%v"?`, contact.Name, contact.Company)
		r.DueTime = time.Now().Add(15 * time.Minute)
		break
	default:
		r.Body = fmt.Sprintf(`Добрый день, %v 👋 Как я понимаю, сейчас Вы трудоустроены в компании "%v". Меня зовут Антон, я - менеджер по набору персонала в компании ITBestTech. Мы хотели бы пригласить Вас на собеседование. Интересно ли Вам наше предложение?`, contact.Name, contact.Company)
		r.DueTime = time.Now().Add(10 * time.Minute)
	}

	return r

}

func (c *TaskDemoServiceImpl) updateLinkedinTask(r *entities.Task) {
	switch r.Action {
	case "view_profile":
		r.Name = "Посмотреть профиль"
		r.Description = "Зайти на страницу профиля Linkedin и подписаться " + r.Contact.Linkedin
		break
	case "private_msg":
		r.Name = "Написать личное сообщение"
		r.Description = "Написать личное сообщение профилю " + r.Contact.Linkedin
		break
	}

	r.Name = "Написать холодное сообщение"
	r.Description = "Зайти на страницу профиля Linkedin и напсать холодное сообщение " + r.Contact.Linkedin

}
