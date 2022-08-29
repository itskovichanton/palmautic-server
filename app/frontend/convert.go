package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"palm/app/entities"
)

func toFrontContact(a *entities.Contact) *Contact {
	r := Contact{}
	copier.Copy(&r, a)
	return &r
}

func toContactModel(a *Contact) *entities.Contact {
	r := entities.Contact{}
	copier.Copy(&r, a)
	return &r
}

func toTaskModel(a *Task) *entities.Task {
	r := entities.Task{}
	copier.Copy(&r, a)
	switch a.Type {
	case TaskType_WRITE_LETTER:
		r.Type = entities.WriteLetter
	}
	switch a.Status {
	case TaskStatus_CLOSED_POSITIVE:
		r.Status = entities.ClosedPositive
	case TaskStatus_CLOSED_NEGATIVE:
		r.Status = entities.ClosedNegative
	case TaskStatus_ACTIVE:
		r.Status = entities.Active
	}
	return &r
}

func toFrontAccount(a *core.Account) *Account {
	return &Account{
		Username: a.Username,
		FullName: a.FullName,
		Id:       a.ID,
		Password: a.Password,
	}
}

func toFrontSession(s *core.Session) *Session {
	return &Session{Token: s.Token}
}

func toBaseError(err *pipeline.Err) *BaseError {
	return &BaseError{
		CommonErrors: getCommonErrors(err),
		Message:      err.Message,
		Reason:       err.Reason,
		Details:      err.Details,
	}
}

func getCommonErrors(err *pipeline.Err) *CommonErrors {
	r := &CommonErrors{}
	empty := true
	switch e := err.Error.(type) {
	case *validation.ValidationError:
		empty = false
		r.ValidationError = &ValidationError{
			Reason:  e.Reason,
			Param:   e.Param,
			Message: e.Message,
			Value: &any.Any{
				Value: []byte(cast.ToString(e.InvalidValue)),
			},
		}
	}
	switch err.Reason {
	case frmclient.ReasonCallerUpdateRequired:
		empty = false
		r.UpdateRequiredError = &UpdateRequiredError{RequiredVersion: 2}
	}
	if empty {
		return nil
	}
	return r
}

func toFrontTask(a *entities.Task) *Task {
	r := Task{}
	copier.Copy(&r, a)
	switch a.Type {
	case entities.WriteLetter:
		r.Type = TaskType_WRITE_LETTER
	}
	switch a.Status {
	case entities.ClosedPositive:
		r.Status = TaskStatus_CLOSED_POSITIVE
	case entities.ClosedNegative:
		r.Status = TaskStatus_CLOSED_NEGATIVE
	case entities.Active:
		r.Status = TaskStatus_ACTIVE
	}
	return &r
}
