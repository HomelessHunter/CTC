package models

import (
	"fmt"
	"testing"

	telegram "github.com/HomelessHunter/CTC/wrapper/models/telegram"
)

func TestUser(t *testing.T) {
	user, err := telegram.NewUser(telegram.WithUserId(-1), telegram.WithUserIsBot(false),
		telegram.WithUserFirstName("Kirill"), telegram.WithUserLastName("Lesnov"), telegram.WithUserUsername("Bob"),
		telegram.WithUserLanguageCode("RU"), telegram.WithUserCanJoinGroups(false), telegram.WithUserCanReadAllGroupMsg(false),
		telegram.WithUserSupportsInlineQueries(false))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(user)
	}
}
