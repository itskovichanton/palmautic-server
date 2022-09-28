package backend

import (
	"fmt"
	"github.com/itskovichanton/core/pkg/core"
	"math/rand"
	"path/filepath"
	"salespalm/server/app/entities"
	"time"
)

type ITaskDemoService interface {
	GenerateTasks(count int, accountId entities.ID) int
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
	templateDir     string
}

func (c *TaskDemoServiceImpl) Init() {
	rand.Seed(42)
	c.templateDir = c.Config.GetOnBaseWorkDir("manual_email_templates")
}

func (c *TaskDemoServiceImpl) FindRandomContact(accountId entities.ID) *entities.Contact {
	return c.ContactService.GetByIndex(accountId, rand.Intn(10000))
}

func (c *TaskDemoServiceImpl) GenerateTasks(count int, accountId entities.ID) int {

	generated := 0
	for i := 0; i < count; i++ {
		contact := c.FindRandomContact(accountId)
		_, err := c.TaskService.CreateOrUpdate(c.generateRandomTask(contact, accountId))
		if err == nil {
			generated++
		}
	}
	return generated
}

func (c *TaskDemoServiceImpl) generateRandomTask(contact *entities.Contact, accountId entities.ID) *entities.Task {

	r := &entities.Task{
		BaseEntity: entities.BaseEntity{
			AccountId: accountId,
		},
		Sequence: &entities.IDAndTitle{
			ID: c.SequenceService.GetByIndex(accountId, 10).GetId(),
		},
		Contact: contact,
	}

	types := c.TaskService.Meta(accountId).Types
	taskType := types[rand.Intn(len(types))]
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
		r.Description = "Написать в личное сообщение Whatsapp: https://wa.me/" + contact.Phone
		break
	case entities.TaskTypeTelegram.Creds.Name:
		r.Name = "Написать в Telegram"
		r.Description = "Написать в личное сообщение Telegram: https://t.me/" + contact.Phone
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
		r.Description = fmt.Sprintf("Отправить письмо для %v (%v)", contact.Name, contact.Phone)
		break
	}

	if r.Type == entities.TaskTypeManualEmail.Creds.Name {
		r.DueTime = time.Now().Add(20 * time.Minute)
		r.Body, _ = c.TemplateService.GetTemplate(filepath.Join(c.templateDir, "manual_email_it_hr.html"), &map[string]interface{}{
			"Contact": r.Contact,
			"Me":      c.AccountService.Accounts()[accountId],
		})
	} else {
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
