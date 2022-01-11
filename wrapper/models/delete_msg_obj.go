package models

type DeleteMsgObj struct {
	ChatId int64 `json:"chat_id"`
	MsgId  int   `json:"message_id"`
}

func NewDeleteMsgObj(chatId int64, msgId int) *DeleteMsgObj {
	return &DeleteMsgObj{chatId, msgId}
}

type DeleteResult struct {
	Ok     bool `json:"ok"`
	Result bool `json:"result"`
}
