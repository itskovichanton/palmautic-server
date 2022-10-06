package entities

import (
	"fmt"
	"time"
)

type Task struct {
	BaseEntity

	Name               string
	Description        string
	Type               string
	Status             string
	StartTime, DueTime time.Time
	Sequence           *IDAndTitle
	Contact            *Contact
	Action             string
	Body               string
	Subject            string
	Alertness          string
}

type TaskCommons struct {
	Types    map[string]*TaskType
	Statuses []string
	Stats    *TaskStats
}

type TaskStats struct {
	All              int
	ByType, ByStatus map[string]int
}

func (t Task) HasFinalStatus() bool {
	return len(t.Status) > 0 && t.Status != TaskStatusPending && t.Status != TaskStatusStarted
}

func (t Task) AutoExecutable() bool {
	return t.Type == TaskTypeAutoEmail.Creds.Name
}

func (t Task) HasTypeEmail() bool {
	return t.Type == TaskTypeManualEmail.Creds.Name || t.Type == TaskTypeAutoEmail.Creds.Name
}

func (t Task) CanExecute() bool {
	return t.IsMessenger() && len(t.Contact.Phone) > 0 || t.HasTypeEmail() && len(t.Contact.Email) > 0 || t.Type == TaskTypeLinkedin.Creds.Name && len(t.Contact.Linkedin) > 0
}

func (t Task) IsMessenger() bool {
	return t.Type != TaskTypeManualEmail.Creds.Name && t.Type != TaskTypeLinkedin.Creds.Name
}

func (t *Task) Refresh() {

	//if len(t.Name) == 0 {
	t.Name = calcName(t)
	//}
	//if len(t.Description) == 0 {
	if t.Contact != nil {
		t.Description = calcDescription(t)
	}
	//}

	calcAlertness(t)
	if !t.HasFinalStatus() {
		calcStatus(t)
	}

}

type TaskType struct {
	Creds   *NameAndTitle
	Actions []*TaskAction
	Order   int
}

type TaskAction NameAndTitle

var (
	TaskTypeAutoEmail = &TaskType{
		Creds: &NameAndTitle{
			Name:  "auto_email",
			Title: "Автоматический Email",
		},
		Actions: []*TaskAction{{
			Name:  "send_letter",
			Title: "Отправить письмо",
		}},
	}

	TaskTypeManualEmail = &TaskType{
		Creds: &NameAndTitle{
			Name:  "manual_email",
			Title: "Мануальный Email",
		},
		Actions: []*TaskAction{{
			Name:  "send_letter",
			Title: "Отправить письмо",
		}},
	}

	TaskTypeWhatsapp = &TaskType{
		Creds: &NameAndTitle{
			Name:  "whatsapp",
			Title: "Whatsapp",
		},
		Actions: []*TaskAction{{
			Name:  "private_msg",
			Title: "Написать личное сообщение",
		}},
	}

	TaskTypeTelegram = &TaskType{
		Creds: &NameAndTitle{
			Name:  "telegram",
			Title: "Telegram",
		},
		Actions: []*TaskAction{{
			Name:  "private_msg",
			Title: "Написать личное сообщение",
		}},
	}

	TaskTypeCall = &TaskType{
		Creds: &NameAndTitle{
			Name:  "call",
			Title: "Звонок",
		},
		Actions: []*TaskAction{{
			Name:  "call",
			Title: "Позвонить",
		}},
	}

	TaskTypeLinkedin = &TaskType{
		Creds: &NameAndTitle{
			Name:  "linkedin",
			Title: "Linkedin",
		},
		Actions: []*TaskAction{{
			Name:  "view_profile",
			Title: "Просмотреть профиль",
		}, {
			Name:  "private_msg",
			Title: "Написать личное сообщение",
		}, {
			Name:  "cold_msg",
			Title: "InMail",
		}, {
			Name:  "connect",
			Title: "Connect",
		}},
	}
)

const (
	TaskStatusCompleted = "completed"
	TaskStatusStarted   = "started"
	TaskStatusSkipped   = "skipped"
	TaskStatusPending   = "pending"
	TaskStatusExpired   = "expired"
	TaskStatusReplied   = "replied"
	TaskStatusArchived  = "archived"

	TaskAlertnessGreen  = "green"
	TaskAlertnessOrange = "orange"
	TaskAlertnessRed    = "red"
	TaskAlertnessGray   = "gray"
	TaskAlertnessBlue   = "blue"
)

func calcLinkedinTaskDescription(t *Task) string {
	switch t.Action {
	case "view_profile":
		return `Зайти на страницу профиля Linkedin и подписаться <a target="_blank" href="{{.Contact.Linkedin}}">{{.Contact.Linkedin}}</a>`
	case "private_msg":
		return `Написать личное сообщение профилю <a target="_blank" href="{{.Contact.Linkedin}}">{{.Contact.Linkedin}}</a>`
	}

	return `Зайти на страницу профиля Linkedin и написать InMail <a target="_blank" href="{{.Contact.Linkedin}}">{{.Contact.Linkedin}}</a>`
}

func calcLinkedinTaskName(t *Task) string {
	switch t.Action {
	case "view_profile":
		return "Посмотреть профиль"
	case "private_msg":
		return "Написать личное сообщение"
	}

	return "Отправить InMail"
}

func calcStatus(t *Task) {
	now := time.Now()
	if t.StartTime.After(now) {
		t.Status = TaskStatusPending
	} else if t.DueTime.Before(now) {
		t.Status = TaskStatusExpired
	} else {
		t.Status = TaskStatusStarted
	}
}

func calcAlertness(t *Task) {
	if t.Status == TaskStatusReplied {
		t.Alertness = TaskAlertnessBlue
	} else if t.HasFinalStatus() {
		t.Alertness = TaskAlertnessGray
	} else {
		durationToDueTime := t.DueTime.Sub(time.Now())
		if durationToDueTime < 0 {
			t.Alertness = TaskAlertnessGray
		} else if durationToDueTime < 5*time.Minute {
			t.Alertness = TaskAlertnessOrange
		} else if durationToDueTime < 2*time.Minute {
			t.Alertness = TaskAlertnessRed
		} else {
			t.Alertness = TaskAlertnessGreen
		}
	}
}

func calcName(t *Task) string {
	switch t.Type {
	case TaskTypeWhatsapp.Creds.Name:
		return "Написать в Whatsapp"
	case TaskTypeTelegram.Creds.Name:
		return "Написать в Telegram"
	case TaskTypeCall.Creds.Name:
		return "Позвонить по телефону"
	case TaskTypeLinkedin.Creds.Name:
		return calcLinkedinTaskName(t)
	case TaskTypeManualEmail.Creds.Name:
		return "Отправить письмо"
	}
	return "Выполнить задачу"
}

func calcDescription(t *Task) string {

	switch t.Type {
	case TaskTypeWhatsapp.Creds.Name:
		return fmt.Sprintf(`Написать в личное сообщение Whatsapp: <a target="_blank" href="%v">{{.Contact.Phone}}</a>`, FormatUrl("https://wa.me", t.Contact.Phone))
	case TaskTypeTelegram.Creds.Name:
		return fmt.Sprintf(`Написать в личное сообщение Telegram: <a target="_blank" href="%v">{{.Contact.Phone}}</a>`, FormatUrl("https://t.me", t.Contact.Phone))
	case TaskTypeCall.Creds.Name:
		return "Позвонить по номеру телефона {{.Contact.Phone}}. Настройся на продуктивный лад 😊"
	case TaskTypeLinkedin.Creds.Name:
		return calcLinkedinTaskDescription(t)
	case TaskTypeManualEmail.Creds.Name:
		return "Отправить письмо для {{.Contact.Name}} на {{.Contact.Email}}"
	}

	return ""
}

func IsTaskAutoExecuted(t *Task) bool {
	return t.Type == TaskTypeAutoEmail.Creds.Name
}
