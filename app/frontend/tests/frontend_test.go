package tests

import (
	"bitbucket.org/itskovich/goava/pkg/goava/httputils"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"palm/app/frontend/grpc_server"
	"testing"
)

func TestLoginUser(t *testing.T) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("127.0.0.1:3001", opts...)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer conn.Close()

	client := grpc_server.NewAccountsClient(conn)
	ctx := context.Background()
	username := "a.itskovich"
	password := "92559255"
	ctx = metadata.AppendToOutgoingContext(ctx, "caller-version-code", "1",
		"caller-version-name", "1.0.0", "caller-type", "tester", "lang", "ru", "authorization", httputils.CalcBasicAuth(username, password))

	r, err := client.Login(ctx, &empty.Empty{})
	if err != nil {
		t.Error(err.Error())
		return
	}

	contactsClient := grpc_server.NewContactsClient(conn)

	contact := &grpc_server.Contact{
		Name:  "Владимир Иванов",
		Email: "v.ivanovmail@gmail.com",
		Phone: "+7929553901",
		Id:    1,
	}
	be, err := contactsClient.Delete(ctx, contact)
	if err != nil {
		t.Error(err.Error())
		return
	}
	println(be)
	//contact = cr.Result

	if r.Error != nil {
		// Обработка ошибок
		commonErrs := r.Error.CommonErrors
		if commonErrs != nil {
			if commonErrs.ValidationError != nil {
				println(fmt.Sprintf("param=%v, value=%v", commonErrs.ValidationError.Param, commonErrs.ValidationError.Value))
			}
			if commonErrs.UpdateRequiredError != nil {
				println(fmt.Sprintf("Требуется обновить приложение до версии %v", commonErrs.UpdateRequiredError.RequiredVersion))
			}
		}
		println(utils.ToJson(r.Error))
	} else {
		println(fmt.Sprintf("Вы зарегестрировались как %v", utils.ToJson(r.Account)))
	}
}
