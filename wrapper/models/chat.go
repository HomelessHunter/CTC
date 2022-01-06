package models

import "errors"

type Chat struct {
	Id          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Bio         string `json:"bio"`
	Description string `json:"description"`
	InviteLink  string `json:"invite_link"`
}

func NewChat(opts ...ChatOption) (*Chat, error) {
	chat := Chat{}

	for _, opt := range opts {
		err := opt(&chat)
		if err != nil {
			return nil, err
		}
	}

	return &chat, nil
}

func (chat *Chat) ID() int64 {
	return chat.Id
}

type ChatOption func(*Chat) error

func WithChatId(id int64) ChatOption {
	return func(c *Chat) error {
		if id < 0 {
			return errors.New("id should be positive")
		}

		c.Id = id
		return nil
	}
}

func WithChatType(t string) ChatOption {
	return func(c *Chat) error {
		if t == "" {
			return errors.New("type shouldn't be empty")
		}

		c.Type = t
		return nil
	}
}

func WithTitle(title string) ChatOption {
	return func(c *Chat) error {
		if title == "" {
			return errors.New("title shoudn't be empty")
		}

		c.Title = title
		return nil
	}
}

func WithChatUsername(username string) ChatOption {
	return func(c *Chat) error {
		if username == "" {
			return errors.New("username shouldn't be empty")
		}

		c.Username = username
		return nil
	}
}

func WithChatFirstName(firstName string) ChatOption {
	return func(c *Chat) error {
		if firstName == "" {
			return errors.New("first name shoudn't be empty")
		}

		c.FirstName = firstName
		return nil
	}
}

func WithChatLastName(lastName string) ChatOption {
	return func(c *Chat) error {
		if lastName == "" {
			return errors.New("last name should't be empty")
		}

		c.LastName = lastName
		return nil
	}
}

func WithChatBio(bio string) ChatOption {
	return func(c *Chat) error {
		if bio == "" {
			return errors.New("bio shoudn't be empty")
		}

		c.Bio = bio
		return nil
	}
}

func WithChatDescription(description string) ChatOption {
	return func(c *Chat) error {
		if description == "" {
			return errors.New("description shouldn't be empty")
		}

		c.Description = description
		return nil
	}
}

func WithChatInviteLink(inviteLink string) ChatOption {
	return func(c *Chat) error {
		if inviteLink == "" {
			return errors.New("invite link shouldn't be empty")
		}

		c.InviteLink = inviteLink
		return nil
	}
}
