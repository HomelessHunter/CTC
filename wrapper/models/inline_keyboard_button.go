package models

import "errors"

type InlineKeyboardButton struct {
	Text              string `json:"text"`
	Url               string `json:"url,omitempty"`
	CallbackData      string `json:"callback_data,omitempty"`
	SwitchInlineQuery string `json:"switch_inline_query,omitempty"`
	SIQCurrentChat    string `json:"switch_inline_query_current_chat,omitempty"`
}

func NewInlineKeyboardButton(opts ...IKBOptions) (*InlineKeyboardButton, error) {
	ikb := InlineKeyboardButton{}

	for _, opt := range opts {
		err := opt(&ikb)
		if err != nil {
			return nil, err
		}
	}

	return &ikb, nil
}

type IKBOptions func(*InlineKeyboardButton) error

func WithIKBText(text string) IKBOptions {
	return func(ikb *InlineKeyboardButton) error {
		if text == "" {
			return errors.New("text shouldn't be empty")
		}

		ikb.Text = text
		return nil
	}
}

func WithIKBUrl(url string) IKBOptions {
	return func(ikb *InlineKeyboardButton) error {
		if url == "" {
			return errors.New("url shouldn't be empty")
		}

		ikb.Url = url
		return nil
	}
}

func WithIKBCallbackData(callbackData string) IKBOptions {
	return func(ikb *InlineKeyboardButton) error {
		if callbackData == "" {
			return errors.New("callback shouldn't be empty")
		}

		ikb.CallbackData = callbackData
		return nil
	}
}

func WithIKBSwitchInlineQuery(switchInlineQuery string) IKBOptions {
	return func(ikb *InlineKeyboardButton) error {
		if switchInlineQuery == "" {
			return errors.New("switchInlineQuery shouldn't be empty")
		}

		ikb.SwitchInlineQuery = switchInlineQuery
		return nil
	}
}

func WithIKBsiqCurrentChat(siqCurrentChat string) IKBOptions {
	return func(ikb *InlineKeyboardButton) error {
		if siqCurrentChat == "" {
			return errors.New("siqCurrentChat shouldn't be empty")
		}

		ikb.SIQCurrentChat = siqCurrentChat
		return nil
	}
}
