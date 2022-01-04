package wrapper

import (
	"context"
	"errors"
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

func NewUser(opts ...UserOption) (*User, error) {
	user := User{}

	for _, opt := range opts {
		err := opt(&user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

type UserOption func(user *User) error

func WithUserId(id int64) UserOption {
	return func(user *User) error {
		if id < 0 {
			return errors.New("id should be positive")
		}
		user.Id = id
		return nil
	}
}

func WithUserIsBot(isBot bool) UserOption {
	return func(user *User) error {
		user.IsBot = isBot
		return nil
	}
}

func WithUserFirstName(firstName string) UserOption {
	return func(user *User) error {
		user.FirstName = firstName
		return nil
	}
}

func WithUserLastName(lastName string) UserOption {
	return func(user *User) error {
		user.LastName = lastName
		return nil
	}
}

func WithUserUsername(username string) UserOption {
	return func(user *User) error {
		user.Username = username
		return nil
	}
}

func WithUserLanguageCode(languageCode string) UserOption {
	return func(user *User) error {
		user.LanguageCode = languageCode
		return nil
	}
}

func WithUserCanJoinGroups(canJoinGroups bool) UserOption {
	return func(user *User) error {
		user.CanJoinGroups = canJoinGroups
		return nil
	}
}

func WithUserCanReadAllGroupMsg(canReadAllGroupMsg bool) UserOption {
	return func(user *User) error {
		user.CanReadAllGroupMsg = canReadAllGroupMsg
		return nil
	}
}

func WithUserSupportsInlineQueries(supportsInlineQueries bool) UserOption {
	return func(user *User) error {
		user.SupportsInlineQueries = supportsInlineQueries
		return nil
	}
}

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
			return errors.New("id shoudl be positive")
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
			return errors.New("entities shouldn't be empty")
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

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

func NewInlineKeyboardMarkup(ik [][]InlineKeyboardButton) (*InlineKeyboardMarkup, error) {
	if len(ik) == 0 {
		return nil, errors.New("ik shoudn't be empty")
	}
	return &InlineKeyboardMarkup{ik}, nil
}

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
