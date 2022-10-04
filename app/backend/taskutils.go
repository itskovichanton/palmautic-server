package backend

import (
	"fmt"
	"salespalm/server/app/entities"
	utils2 "salespalm/server/app/utils"
	"time"
)

func RefreshTask(t *entities.Task) {

	//if len(t.Name) == 0 {
	t.Name = calcName(t)
	//}
	//if len(t.Description) == 0 {
	t.Description = calcDescription(t)
	//}

	if t.HasFinalStatus() {
		return
	}

	calcStatus(t)
	calcAlertness(t)

}

func calcLinkedinTaskDescription(t *entities.Task) string {
	switch t.Action {
	case "view_profile":
		return "Зайти на страницу профиля Linkedin и подписаться {{.Contact.Linkedin}}"
	case "private_msg":
		return "Написать личное сообщение профилю {{.Contact.Linkedin}}"
	}

	return "Зайти на страницу профиля Linkedin и написать холодное сообщение {{.Contact.Linkedin}}"
}

func calcLinkedinTaskName(t *entities.Task) string {
	switch t.Action {
	case "view_profile":
		return "Посмотреть профиль"
	case "private_msg":
		return "Написать личное сообщение"
	}

	return "Отправить InMail"
}

func calcStatus(t *entities.Task) {
	now := time.Now()
	if t.StartTime.After(now) {
		t.Status = entities.TaskStatusPending
	} else if t.DueTime.Before(now) {
		t.Status = entities.TaskStatusExpired
	} else {
		t.Status = entities.TaskStatusStarted
	}
}

func calcAlertness(t *entities.Task) {
	if t.HasFinalStatus() {
		t.Alertness = entities.TaskAlertnessGray
	} else {
		durationToDueTime := t.DueTime.Sub(time.Now())
		if durationToDueTime < 0 {
			t.Alertness = entities.TaskAlertnessGray
		} else if durationToDueTime < 5*time.Minute {
			t.Alertness = entities.TaskAlertnessOrange
		} else if durationToDueTime < 2*time.Minute {
			t.Alertness = entities.TaskAlertnessRed
		} else {
			t.Alertness = entities.TaskAlertnessGreen
		}
	}
}

func calcName(t *entities.Task) string {
	switch t.Type {
	case entities.TaskTypeWhatsapp.Creds.Name:
		return "Написать в Whatsapp"
	case entities.TaskTypeTelegram.Creds.Name:
		return "Написать в Telegram"
	case entities.TaskTypeCall.Creds.Name:
		return "Позвонить по телефону"
	case entities.TaskTypeLinkedin.Creds.Name:
		return calcLinkedinTaskName(t)
	case entities.TaskTypeManualEmail.Creds.Name:
		return "Отправить письмо"
	}
	return "Выполнить задачу"
}

func calcDescription(t *entities.Task) string {

	switch t.Type {
	case entities.TaskTypeWhatsapp.Creds.Name:
		return fmt.Sprintf(`Написать в личное сообщение Whatsapp: <a target="_blank" href="%v">{{.Contact.Phone}}</a>`, utils2.FormatUrl("https://wa.me", t.Contact.Phone))
	case entities.TaskTypeTelegram.Creds.Name:
		return fmt.Sprintf(`Написать в личное сообщение Telegram: <a target="_blank" href="%v">{{.Contact.Phone}}</a>`, utils2.FormatUrl("https://t.me", t.Contact.Phone))
	case entities.TaskTypeCall.Creds.Name:
		return "Позвонить по номеру телефона {{.Contact.Phone}}. Настройся на продуктивный лад 😊"
	case entities.TaskTypeLinkedin.Creds.Name:
		return calcLinkedinTaskDescription(t)
	case entities.TaskTypeManualEmail.Creds.Name:
		return "Отправить письмо для {{.Contact.Name}} на {{.Contact.Email}}"
	}

	return ""
}

func IsTaskAutoExecuted(t *entities.Task) bool {
	return t.Type == entities.TaskTypeAutoEmail.Creds.Name
}
