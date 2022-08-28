package backend

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"encoding/json"
	"os"
	"palm/app/backend"
	"testing"
)

func TestFillDb(t *testing.T) {
	r := backend.DBContent{
		Accounts: map[int]*core.Account{
			1001: {
				ID:           1001,
				Username:     "a.itskovich",
				Lang:         "ru",
				FullName:     "Ицкович Антон Евгеньевич",
				SessionToken: "user-1",
				Role:         "user",
				Password:     "92559255",
			},
			1002: {
				ID:           1002,
				Username:     "shlomo",
				Lang:         "ru",
				FullName:     "Шломо",
				SessionToken: "user-2",
				Role:         "user",
				Password:     "f92559255dfs",
			}},
	}

	dataBytes, _ := json.Marshal(&r)
	os.WriteFile("db.json", dataBytes, 0644)
}
