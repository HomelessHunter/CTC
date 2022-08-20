package models

import "errors"

type WHUpdate struct {
	Ok     bool   `json:"ok"`
	Result Update `json:"result"`
}

func NewWhUpdate() *WHUpdate {
	return &WHUpdate{}
}

type TGUpdate struct {
	IsOK   bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	Id             int           `json:"update_id"`
	Msg            Message       `json:"message"`
	EditedMsg      Message       `json:"edited_message"`
	ChanPost       Message       `json:"channel_post"`
	EditedChanPost Message       `json:"edited_channel_post"`
	InlineQuery    InlineQuery   `json:"inline_query"`
	CallbackQuery  CallbackQuery `json:"callback_query"`
}

func NewUpdate(opts ...UpdateOpts) (*Update, error) {
	update := Update{}

	for _, opt := range opts {
		err := opt(&update)
		if err != nil {
			return nil, err
		}
	}

	return &update, nil
}

func (update *Update) FromUser() *User {
	return &update.Msg.From
}

func (update *Update) FromChat() *Chat {
	return &update.Msg.Chat
}

func (update *Update) GetCallbackData() string {
	return update.CallbackQuery.Data
}

// func (update *Update) Update

type UpdateOpts func(*Update) error

func WithUpdateId(id int) UpdateOpts {
	return func(u *Update) error {
		if id < 0 {
			return errors.New("id should be positive")
		}

		u.Id = id
		return nil
	}
}

func WithUpdateMsg(msg *Message) UpdateOpts {
	return func(u *Update) error {
		if msg == nil {
			return errors.New("msg shouldn't be empty")
		}

		u.Msg = *msg
		return nil
	}
}

func WithUpdateEditedMsg(editedMsg *Message) UpdateOpts {
	return func(u *Update) error {
		if editedMsg == nil {
			return errors.New("editedMsg shouldn't be empty")
		}

		u.EditedMsg = *editedMsg
		return nil
	}
}

func WithUpdateChanPost(chanPost *Message) UpdateOpts {
	return func(u *Update) error {
		if chanPost == nil {
			return errors.New("chanPost shouldn't be empty")
		}

		u.ChanPost = *chanPost
		return nil
	}
}

func WithUpdateEditedChanPost(editedChanPost *Message) UpdateOpts {
	return func(u *Update) error {
		if editedChanPost == nil {
			return errors.New("editedChanPost shouldn't be empty")
		}

		u.EditedChanPost = *editedChanPost
		return nil
	}
}

func WithUpdateInlineQuery(inlineQuery *InlineQuery) UpdateOpts {
	return func(u *Update) error {
		if inlineQuery == nil {
			return errors.New("inlineQuery shouldn't be empty")
		}

		u.InlineQuery = *inlineQuery
		return nil
	}
}

func WithUpdateCallbackQuery(callbackQuery *CallbackQuery) UpdateOpts {
	return func(u *Update) error {
		if callbackQuery == nil {
			return errors.New("callbackQuery shouldn't be empty")
		}

		u.CallbackQuery = *callbackQuery
		return nil
	}
}
