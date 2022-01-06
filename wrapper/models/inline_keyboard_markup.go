package models

import "errors"

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

func NewInlineKeyboardMarkup(ik [][]InlineKeyboardButton) (*InlineKeyboardMarkup, error) {
	if len(ik) == 0 {
		return nil, errors.New("ik shoudn't be empty")
	}
	return &InlineKeyboardMarkup{ik}, nil
}
