package models

import "errors"

type MessageEntity struct {
	Type     string `json:"type"`
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	Url      string `json:"url"`
	User     User   `json:"user"`
	Language string `json:"language"`
}

func NewMsgEntity(opts ...MsgEntityOptions) (*MessageEntity, error) {
	msgEntity := MessageEntity{}

	for _, opt := range opts {
		err := opt(&msgEntity)
		if err != nil {
			return nil, err
		}
	}

	return &msgEntity, nil
}

type MsgEntityOptions func(*MessageEntity) error

func WithMEType(t string) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if t == "" {
			return errors.New("type shouldn't be empty")
		}

		me.Type = t
		return nil
	}
}

func WithMEOffset(offset int) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if offset == 0 {
			return errors.New("offset shouldn't be 0")
		}

		me.Offset = offset
		return nil
	}
}

func WithMELength(length int) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if length <= 0 {
			return errors.New("length shoudn't be empty")
		}

		me.Length = length
		return nil
	}
}

func WithMEUrl(url string) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if url == "" {
			return errors.New("url shouldn't be empty")
		}

		me.Url = url
		return nil
	}
}

func WithMEUser(user *User) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if user == nil {
			return errors.New("user shouldn't be empty")
		}

		me.User = *user
		return nil
	}
}

func WithMELanguage(language string) MsgEntityOptions {
	return func(me *MessageEntity) error {
		if language == "" {
			return errors.New("language shouldn't be empty")
		}

		me.Language = language
		return nil
	}
}
