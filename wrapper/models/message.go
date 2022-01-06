package models

import "errors"

type Message struct {
	Id          int                  `json:"message_id,omitempty"`
	From        User                 `json:"from,omitempty"`
	SenderChat  Chat                 `json:"sender_chat,omitempty"`
	Date        int                  `json:"date,omitempty"`
	Chat        Chat                 `json:"chat,omitempty"`
	Text        string               `json:"text,omitempty"`
	Entities    []MessageEntity      `json:"entities,omitempty"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func NewMsg(opts ...MsgOptions) (*Message, error) {
	msg := Message{}

	for _, opt := range opts {
		err := opt(&msg)
		if err != nil {
			return nil, err
		}
	}

	return &msg, nil
}

func (msg *Message) FromChatID() int64 {
	return msg.Chat.Id
}

type MsgOptions func(*Message) error

func WithMsgId(id int) MsgOptions {
	return func(m *Message) error {
		if id < 0 {
			return errors.New("id should be positive")
		}

		m.Id = id
		return nil
	}
}

func WithMsgFrom(from *User) MsgOptions {
	return func(m *Message) error {
		if from == nil {
			return errors.New("from shouldn't be empty")
		}

		m.From = *from
		return nil
	}
}

func WithMsgSenderChat(senderChat *Chat) MsgOptions {
	return func(m *Message) error {
		if senderChat == nil {
			return errors.New("senderChat shouldn't be empty")
		}

		m.SenderChat = *senderChat
		return nil
	}
}

func WithMsgDate(date int) MsgOptions {
	return func(m *Message) error {
		if date <= 0 {
			return errors.New("date should be greater than 0")
		}

		m.Date = date
		return nil
	}
}

func WithMsgChat(chat *Chat) MsgOptions {
	return func(m *Message) error {
		if chat == nil {
			return errors.New("chat shouldn't be empty")
		}

		m.Chat = *chat
		return nil
	}
}

func WithMsgText(text string) MsgOptions {
	return func(m *Message) error {
		if text == "" {
			return errors.New("text shouldn't be empty")
		}

		m.Text = text
		return nil
	}
}

func WithMsgEntities(entities []MessageEntity) MsgOptions {
	return func(m *Message) error {
		if len(entities) == 0 {
			return errors.New("entities shoildn't be empty")
		}

		m.Entities = entities
		return nil
	}
}

func WithMsgReplyMarkup(replyMarkup *InlineKeyboardMarkup) MsgOptions {
	return func(m *Message) error {
		if replyMarkup == nil {
			return errors.New("replyMarkup shouldn't be empty")
		}

		m.ReplyMarkup = *replyMarkup
		return nil
	}
}
