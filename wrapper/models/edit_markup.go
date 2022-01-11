package models

import "errors"

type EditMarkupObj struct {
	ChatId      int64                `json:"chat_id,omitempty"`
	MsgId       int                  `json:"message_id,omitempty"`
	InlineMsgId int                  `json:"inline_message_id,omitempty"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func NewEditMSGReplyMarkup(opts ...EditMarkupObjOpts) (*EditMarkupObj, error) {
	editMarkup := EditMarkupObj{}

	for _, opt := range opts {
		err := opt(&editMarkup)
		if err != nil {
			return nil, err
		}
	}

	return &editMarkup, nil
}

type EditMarkupObjOpts func(*EditMarkupObj) error

func WithEMOChatID(chatId int64) EditMarkupObjOpts {
	return func(emo *EditMarkupObj) error {
		if chatId < 0 {
			return errors.New("chatId should be positive")
		}

		emo.ChatId = chatId
		return nil
	}
}

func WithEMOMsgID(msgId int) EditMarkupObjOpts {
	return func(emo *EditMarkupObj) error {
		if msgId < 0 {
			return errors.New("msgId should be positive")
		}

		emo.MsgId = msgId
		return nil
	}
}

func WithEMOInlineMsgID(inlineId int) EditMarkupObjOpts {
	return func(emo *EditMarkupObj) error {
		if inlineId < 0 {
			return errors.New("inlineId should be positive")
		}

		emo.InlineMsgId = inlineId
		return nil
	}
}

func WithEMOReplyMarkup(replyMarkup *InlineKeyboardMarkup) EditMarkupObjOpts {
	return func(emo *EditMarkupObj) error {
		if replyMarkup == nil {
			return errors.New("replyMarkup shouldn't be empty")
		}

		emo.ReplyMarkup = *replyMarkup
		return nil
	}
}
