package models

import "errors"

type User struct {
	Id                    int64  `json:"id"`
	IsBot                 bool   `json:"is_bot"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	Username              string `json:"username"`
	LanguageCode          string `json:"language_code"`
	CanJoinGroups         bool   `json:"can_join_groups"`
	CanReadAllGroupMsg    bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries bool   `json:"supports_inline_queries"`
}

func NewUser(opts ...UserOption) (*User, error) {
	user := User{}

	for _, opt := range opts {
		err := opt(&user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (user *User) ID() int64 {
	return user.Id
}

type UserOption func(user *User) error

func WithUserId(id int64) UserOption {
	return func(user *User) error {
		if id < 0 {
			return errors.New("id should be positive")
		}
		user.Id = id
		return nil
	}
}

func WithUserIsBot(isBot bool) UserOption {
	return func(user *User) error {
		user.IsBot = isBot
		return nil
	}
}

func WithUserFirstName(firstName string) UserOption {
	return func(user *User) error {
		if firstName == "" {
			return errors.New("firstName shouldn't be empty")
		}

		user.FirstName = firstName
		return nil
	}
}

func WithUserLastName(lastName string) UserOption {
	return func(user *User) error {
		if lastName == "" {
			return errors.New("lastName shouldn't be empty")
		}

		user.LastName = lastName
		return nil
	}
}

func WithUserUsername(username string) UserOption {
	return func(user *User) error {
		if username == "" {
			return errors.New("username shouldn't be empty")
		}

		user.Username = username
		return nil
	}
}

func WithUserLanguageCode(languageCode string) UserOption {
	return func(user *User) error {
		if languageCode == "" {
			return errors.New("languageCode shouldn't be empty")
		}

		user.LanguageCode = languageCode
		return nil
	}
}

func WithUserCanJoinGroups(canJoinGroups bool) UserOption {
	return func(user *User) error {
		user.CanJoinGroups = canJoinGroups
		return nil
	}
}

func WithUserCanReadAllGroupMsg(canReadAllGroupMsg bool) UserOption {
	return func(user *User) error {
		user.CanReadAllGroupMsg = canReadAllGroupMsg
		return nil
	}
}

func WithUserSupportsInlineQueries(supportsInlineQueries bool) UserOption {
	return func(user *User) error {
		user.SupportsInlineQueries = supportsInlineQueries
		return nil
	}
}
