package models

import "errors"

type CallbackQuery struct {
	Id           string  `json:"id"`
	From         User    `json:"from"`
	Msg          Message `json:"message"`
	InlineMsgId  string  `json:"inline_message_id"`
	ChatInstance string  `json:"chat_instance"`
	Data         string  `json:"data"`
}

func NewCallbackQuery(opts ...CallbackQueryOpts) (*CallbackQuery, error) {
	callbackQuery := CallbackQuery{}

	for _, opt := range opts {
		err := opt(&callbackQuery)
		if err != nil {
			return nil, err
		}
	}

	return &callbackQuery, nil
}

func (callbackQuery *CallbackQuery) GetData() string {
	return callbackQuery.Data
}

func (callbackQuery *CallbackQuery) SetData(data string) {
	callbackQuery.Data = data
}

type CallbackQueryOpts func(*CallbackQuery) error

func WithCallbackId(id string) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if id == "" {
			return errors.New("id shouldn't be empty")
		}

		cq.Id = id
		return nil
	}
}

func WithCallbackFrom(from *User) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if from == nil {
			return errors.New("from shouldn't be empty")
		}

		cq.From = *from
		return nil
	}
}

func WithCallbackMsg(editedMsg *Message) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if editedMsg == nil {
			return errors.New("msg shouldn't be empty")
		}

		cq.Msg = *editedMsg
		return nil
	}
}

func WithCallbackInlineMsgId(inlineMsgId string) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if inlineMsgId == "" {
			return errors.New("inlineMsgId shouldn't be empty")
		}

		cq.InlineMsgId = inlineMsgId
		return nil
	}
}

func WithCallbackChatInst(chatInstance string) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if chatInstance == "" {
			return errors.New("chatInstance shouldn't be empty")
		}

		cq.ChatInstance = chatInstance
		return nil
	}
}

func WithCallbackData(data string) CallbackQueryOpts {
	return func(cq *CallbackQuery) error {
		if data == "" {
			return errors.New("data shouldn't be empty")
		}

		cq.Data = data
		return nil
	}
}
