package models

import (
	"fmt"
	"testing"
)

func TestUser(t *testing.T) {
	user, err := NewUser(WithUserId(-1), WithUserIsBot(false),
		WithUserFirstName("Kirill"), WithUserLastName("Lesnov"), WithUserUsername("Bob"),
		WithUserLanguageCode("RU"), WithUserCanJoinGroups(false), WithUserCanReadAllGroupMsg(false),
		WithUserSupportsInlineQueries(false))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(user)
	}
}
