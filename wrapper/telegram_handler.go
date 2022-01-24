package wrapper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	other "github.com/HomelessHunter/CTC/wrapper/models/other"
	telegram "github.com/HomelessHunter/CTC/wrapper/models/telegram"
)

var ErrEmptyPairs = errors.New("pairs shouldn't be empty")

func SetWebhook(client *http.Client) {
	apiKey := os.Getenv("TG")
	postBody, err := json.Marshal(map[string]string{
		"url": fmt.Sprintf("http://localhost:8080/%s", apiKey),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	body := bytes.NewBuffer(postBody)
	resp, err := client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", apiKey), "application/json", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Fprintln(os.Stdout, string(data))
}

func CommandRouter(command string, regs map[string]*regexp.Regexp,
	update *telegram.Update, client *http.Client, pairs []string) (*other.WSQuery, string, error) {

	switch {
	case regs["start"].MatchString(command):
		msg, err := telegram.NewMsg(telegram.WithMsgText("Hello"), telegram.WithMsgChat(update.FromChat()))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, "", err
		}

		sendMsg(client, msg, false)

	case regs["alert"].MatchString(command):
		c := regs["splitter"].Split(command, 3)
		price, err := strconv.ParseFloat(c[2], 32)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, "", err
		}
		market, err := GetMarket(c[1], client)
		if err != nil {
			return nil, "", err
		}

		wsQuery, err := other.NewWsQuery(
			other.WithWSUserId(update.FromUser().ID()),
			other.WithWSChatId(update.FromChat().ID()),
			other.WithWSMarket(market),
			other.WithWSPair(c[1]),
			other.WithWSPrice(float32(price)),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, "", err
		}

		return wsQuery, "alert", nil

	case regs["disconnect"].MatchString(command):

		ik, err := composeKeyboardMarkup(pairs)
		if err != nil {

			if err == ErrEmptyPairs {
				msg, err := telegram.NewMsg(
					telegram.WithMsgId(update.Id),
					telegram.WithMsgChat(update.FromChat()),
					telegram.WithMsgText("You have no alerts"),
				)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return nil, "", err
				}
				sendNDiscardMsg(client, msg, false, 2)
				return nil, "", err
			}

			fmt.Fprintln(os.Stderr, err)
			return nil, "", err
		}
		sendDisconnectMsg(client, update.FromChat(), ik)

	}
	return nil, "", fmt.Errorf("no such command %s", command)
}

func CallbackHandler(client *http.Client, update *telegram.Update, regs map[string]*regexp.Regexp) (*telegram.CallbackQuery, string, error) {
	switch {
	case regs["disconnect"].MatchString(update.GetCallbackData()):
		fmt.Println("CallbackHandler", update.CallbackQuery.From)
		callbackData := regs["splitter"].Split(update.GetCallbackData(), 2)[1]
		market, err := GetMarket(callbackData, client)
		if err != nil {
			return nil, "", err
		}
		update.CallbackQuery.SetData(fmt.Sprintf("%s %s", callbackData, market))

		return &update.CallbackQuery, "disconnect", nil

	default:

	}
	return nil, "", fmt.Errorf("no handler for such callback %s", update.GetCallbackData())
}

func SplitCallbackData(data string) (pair string, market string) {
	callback := strings.Split(data, " ")
	pair = callback[0]
	market = callback[1]
	return
}

func SendCallbackAnswer(client *http.Client, callbackAnswer *telegram.CallbackAnswer) {
	data, err := json.Marshal(callbackAnswer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	body := bytes.NewReader(data)
	_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", os.Getenv("TG")), "application/json", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func sendMsg(client *http.Client, msg *telegram.Message, notify bool) *telegram.Message {
	sendObj, err := telegram.NewSendMsgObj(
		telegram.WithSendChatId(msg.FromChatID()),
		telegram.WithSendText(msg.Text),
		telegram.WithSendParseMode("HTML"),
		telegram.WithSendDisableNotification(notify),
		telegram.WithSendAllowReply(true),
		telegram.WithSendReplyMarkup(msg.ReplyMarkup),
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}

	data, err := json.Marshal(sendObj)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg:", err)
		return nil
	}

	body := bytes.NewReader(data)
	resp, err := client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TG")), "application/json", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg:", err)
		return nil
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg:", err)
		return nil
	}
	defer resp.Body.Close()

	fmt.Println(string(data))

	responseMsg := telegram.NewResponseMessage()

	err = json.Unmarshal(data, responseMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg:", err)
		return nil
	}

	fmt.Println(responseMsg.Result)

	return &responseMsg.Result
}

func sendNDiscardMsg(client *http.Client, msg *telegram.Message, notify bool, cacheTimer int) {
	respMsg := sendMsg(client, msg, notify)
	<-time.After(time.Duration(cacheTimer) * time.Second)
	deleteMsg(client, respMsg.FromChatID(), respMsg.Id)
}

func sendDisconnectMsg(client *http.Client, chat *telegram.Chat, ik *telegram.InlineKeyboardMarkup) {

	msg, err := telegram.NewMsg(
		telegram.WithMsgChat(chat),
		telegram.WithMsgText("Choose alerts to disconnect"),
		telegram.WithMsgReplyMarkup(ik),
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	sendMsg(client, msg, false)
}

func EditMarkup(client *http.Client, callback *telegram.CallbackQuery, pairs []string) error {
	if len(pairs) == 0 {
		// delete msg
		if !deleteMsg(client, callback.Msg.FromChatID(), callback.Msg.Id) {
			return errors.New("could not delete a message")
		}
		return nil
	}
	ik, err := composeKeyboardMarkup(pairs)
	if err != nil {
		return err
	}

	editMarkup, err := telegram.NewEditMSGReplyMarkup(
		telegram.WithEMOChatID(callback.Msg.FromChatID()),
		telegram.WithEMOMsgID(callback.Msg.Id),
		telegram.WithEMOReplyMarkup(ik),
	)
	if err != nil {
		return err
	}

	editMSGReplyMarkup(client, editMarkup)
	return nil
}

func editMSGReplyMarkup(client *http.Client, editMarkup *telegram.EditMarkupObj) {
	data, err := json.Marshal(editMarkup)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	body := bytes.NewReader(data)
	_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/editMessageReplyMarkup", os.Getenv("TG")), "application/json", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func composeKeyboardMarkup(pairs []string) (*telegram.InlineKeyboardMarkup, error) {
	if len(pairs) > 0 {
		inlineButtons := make([]telegram.InlineKeyboardButton, len(pairs)+1)

		disconnectAllBut, err := telegram.NewInlineKeyboardButton(
			telegram.WithIKBText("All"),
			telegram.WithIKBCallbackData("disconnect all"),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}

		inlineButtons[len(inlineButtons)-1] = *disconnectAllBut

		for i, v := range pairs {
			ikb, err := telegram.NewInlineKeyboardButton(
				telegram.WithIKBText(v),
				telegram.WithIKBCallbackData(fmt.Sprintf("disconnect %s", v)),
			)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil, err
			}

			inlineButtons[i] = *ikb
		}

		size := 3

		switch {
		case (len(inlineButtons) % size) == 0:
			size = len(inlineButtons) / size
		default:
			size = len(inlineButtons)/size + 1
		}

		inlineKeyboard := make([][]telegram.InlineKeyboardButton, size)

		for i := range inlineKeyboard {
			size = 3
			if len(inlineButtons) <= 3 {
				size = len(inlineButtons)
			}
			inlineKeyboard[i], inlineButtons = inlineButtons[:size], inlineButtons[size:]
		}

		ik, err := telegram.NewInlineKeyboardMarkup(inlineKeyboard)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}

		fmt.Println(inlineKeyboard)

		return ik, nil
	} else {
		return nil, ErrEmptyPairs
	}

}

func deleteMsg(client *http.Client, chatID int64, msgID int) bool {
	data, err := json.Marshal(telegram.NewDeleteMsgObj(chatID, msgID))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	resp, err := client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", os.Getenv("TG")), "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	defer resp.Body.Close()

	ok := &telegram.DeleteResult{}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = json.Unmarshal(data, ok)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	return ok.Ok
}
