package wrapper

import (
	"context"
)

// Crypto

type Ticker struct {
	Stream string                 `json:"-"`
	Data   map[string]interface{} `json:"data"`
}

func (ticker *Ticker) GetLastPrice() interface{} {
	return ticker.Data["c"]
}

type AvgPrice struct {
	Mins  int    `json:"mins"`
	Price string `json:"price"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}

//

// Telegram

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

type Chat struct {
	Id           int64  `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Bio          string `json:"bio"`
	Descriptions string `json:"description"`
	InviteLink   string `json:"invite_link"`
}

type Message struct {
	Id          int                  `json:"message_id"`
	From        User                 `json:"from"`
	SenderChat  Chat                 `json:"sender_chat"`
	Date        int                  `json:"date"`
	Chat        Chat                 `json:"chat"`
	Text        string               `json:"text"`
	Entities    []MessageEntity      `json:"entities"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup"`
}

type MessageEntity struct {
	Type     string `json:"type"`
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	Url      string `json:"url"`
	User     User   `json:"user"`
	Language string `json:"language"`
}

type SendMsgObj struct {
	ChatId                int64                `json:"chat_id,omitempty"`
	Text                  string               `json:"text,omitempty"`
	ParseMode             string               `json:"parse_mode,omitempty"`
	Entities              []MessageEntity      `json:"entities,omitempty"`
	DisableWebPreview     bool                 `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool                 `json:"disable_notification,omitempty"`
	ReplyToMsgId          int                  `json:"reply_to_message_id,omitempty"`
	AllowSendWithoutReply bool                 `json:"allow_sending_without_reply,omitempty"`
	ReplyMarkup           InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// Don't really need that right now
type InlineQuery struct {
	Id       string `json:"id"`
	From     User   `json:"from"`
	Query    string `json:"query"`
	Offset   string `json:"offset"`
	ChatType string `json:"chat_type"`
}

type CallbackQuery struct {
	Id           string  `json:"id"`
	From         User    `json:"from"`
	Msg          Message `json:"message"`
	InlineMsgId  string  `json:"inline_message_id"`
	ChatInstance string  `json:"chat_instance"`
	Data         string  `json:"data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

type InlineKeyboardButton struct {
	Text              string `json:"text"`
	Url               string `json:"url,omitempty"`
	CallbackData      string `json:"callback_data,omitempty"`
	SwitchInlineQuery string `json:"switch_inline_query,omitempty"`
	SIQCurrentChat    string `json:"switch_inline_query_current_chat,omitempty"`
}

//

type WSQuery struct {
	UserId int64   `json:"user_id"`
	ChatId int64   `json:"chat_id"`
	Pair   string  `json:"pair"`
	Price  float64 `json:"price"`
}

type UserStreams struct {
	ChatId int64
	Pairs  []string
	Cancel context.CancelFunc
	// ShutdownCh is giving a signal to close current websocket connection completely
	ShutdownCh chan int
	// ReconnectCh is giving a signal to close current connection and reconnect with new values
	ReconnectCh chan int
}
