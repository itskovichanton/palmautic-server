package frontend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/frmclient"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/server/pkg/server/pipeline"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/spf13/cast"
	grpc_services "molbulak-apps-golang/telephony/app/frontend/grpc-services"
)

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

func toSession(s *core.Session) *grpc_services.Session {
	return &grpc_services.Session{
		Account: &grpc_services.Account{
			Username: s.Account.Username,
			FullName: s.Account.FullName,
			Lang:     s.Account.Lang,
			Role:     s.Account.Role,
			Id:       s.Account.MCLID,
		},
		Token: s.Token,
	}
}
