package models

import "errors"

type SendMsgObj struct {
	ChatId                int64                `json:"chat_id"`
	Text                  string               `json:"text,omitempty"`
	ParseMode             string               `json:"parse_mode,omitempty"`
	Entities              []MessageEntity      `json:"entities,omitempty"`
	DisableWebPreview     bool                 `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool                 `json:"disable_notification,omitempty"`
	ReplyToMsgId          int                  `json:"reply_to_message_id,omitempty"`
	AllowSendWithoutReply bool                 `json:"allow_sending_without_reply,omitempty"`
	ReplyMarkup           InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func NewSendMsgObj(opts ...SendMsgObjOpts) (*SendMsgObj, error) {
	smo := SendMsgObj{}

	for _, opt := range opts {
		err := opt(&smo)
		if err != nil {
			return nil, err
		}
	}

	return &smo, nil
}

type SendMsgObjOpts func(*SendMsgObj) error

func WithSendChatId(chatId int64) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		if chatId < 0 {
			return errors.New("id should be positive")
		}

		smo.ChatId = chatId
		return nil
	}
}

func WithSendText(text string) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		if text == "" {
			return errors.New("text shouldn't be empty")
		}

		smo.Text = text
		return nil
	}
}

func WithSendParseMode(parseMode string) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		if parseMode == "" {
			return errors.New("parseMode shouldn't be empty")
		}

		smo.ParseMode = parseMode
		return nil
	}
}

func WithSendEntities(entities []MessageEntity) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		if len(entities) == 0 {
			return errors.New("sendObj entities shouldn't be empty")
		}

		smo.Entities = entities
		return nil
	}
}

func WithSendDisableWebPrev(disableWebPreview bool) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		smo.DisableWebPreview = disableWebPreview
		return nil
	}
}

func WithSendDisableNotification(disableNotification bool) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		smo.DisableNotification = disableNotification
		return nil
	}
}

func WithSendReplyToMsgId(replyToMsgId int) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		if replyToMsgId < 0 {
			return errors.New("replyToMsgId shouldn't be empty")
		}

		smo.ReplyToMsgId = replyToMsgId
		return nil
	}
}

func WithSendAllowReply(allowSendWithoutReply bool) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		smo.AllowSendWithoutReply = allowSendWithoutReply
		return nil
	}
}

func WithSendReplyMarkup(replyMarkup InlineKeyboardMarkup) SendMsgObjOpts {
	return func(smo *SendMsgObj) error {
		smo.ReplyMarkup = replyMarkup
		return nil
	}
}
