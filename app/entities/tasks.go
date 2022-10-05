package entities

import (
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

func (t Task) HasTypeEmail() bool {
	return t.Type == TaskTypeManualEmail.Creds.Name || t.Type == TaskTypeAutoEmail.Creds.Name
}

func (t Task) CanExecute() bool {
	return t.IsMessenger() && len(t.Contact.Phone) > 0 || t.HasTypeEmail() && len(t.Contact.Email) > 0 || t.Type == TaskTypeLinkedin.Creds.Name && len(t.Contact.Linkedin) > 0
}

func (t Task) IsMessenger() bool {
	return t.Type != TaskTypeManualEmail.Creds.Name && t.Type != TaskTypeLinkedin.Creds.Name
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
)
