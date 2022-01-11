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
	"time"

	"github.com/HomelessHunter/CTC/wrapper/models"
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
	update *models.Update, client *http.Client, pairs []string) (*models.WSQuery, string) {

	switch {
	case regs["start"].MatchString(command):
		msg, err := models.NewMsg(models.WithMsgText("Hello"), models.WithMsgChat(update.FromChat()))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, ""
		}

		sendMsg(client, msg, false)

	case regs["alert"].MatchString(command):
		c := regs["splitter"].Split(command, 3)
		price, err := strconv.ParseFloat(c[2], 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, ""
		}

		wsQuery, err := models.NewWsQuery(
			models.WithWSUserId(update.FromUser().ID()),
			models.WithWSChatId(update.FromChat().ID()),
			models.WithWSPair(c[1]),
			models.WithWSPrice(price),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, ""
		}

		return wsQuery, "alert"

	case regs["disconnect"].MatchString(command):

		ik, err := composeKeyboardMarkup(pairs)
		if err != nil {

			if err == ErrEmptyPairs {
				msg, err := models.NewMsg(
					models.WithMsgId(update.Id),
					models.WithMsgChat(update.FromChat()),
					models.WithMsgText("You have no alerts"),
				)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return nil, ""
				}
				sendNDiscardMsg(client, msg, false, 2)
				return nil, "disconnect"
			}

			fmt.Fprintln(os.Stderr, err)
			return nil, "disconnect"
		}
		sendDisconnectMsg(client, update.FromChat(), ik)

	}
	return nil, ""
}

func CallbackHandler(update *models.Update, regs map[string]*regexp.Regexp) (*models.CallbackQuery, string) {
	switch {
	case regs["disconnect"].MatchString(update.GetCallbackData()):
		fmt.Println("CallbackHandler", update.CallbackQuery.From)
		callbackData := regs["splitter"].Split(update.GetCallbackData(), 2)[1]
		update.CallbackQuery.SetData(callbackData)

		return &update.CallbackQuery, "disconnect"

	default:

	}
	return nil, ""
}

func SendCallbackAnswer(client *http.Client, callbackAnswer *models.CallbackAnswer) {
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

func sendMsg(client *http.Client, msg *models.Message, notify bool) *models.Message {
	sendObj, err := models.NewSendMsgObj(
		models.WithSendChatId(msg.FromChatID()),
		models.WithSendText(msg.Text),
		models.WithSendParseMode("HTML"),
		models.WithSendDisableNotification(notify),
		models.WithSendAllowReply(true),
		models.WithSendReplyMarkup(msg.ReplyMarkup),
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

	responseMsg := models.NewResponseMessage()

	err = json.Unmarshal(data, responseMsg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg:", err)
		return nil
	}

	fmt.Println(responseMsg.Result)

	return &responseMsg.Result
}

func sendNDiscardMsg(client *http.Client, msg *models.Message, notify bool, cacheTimer int) {
	respMsg := sendMsg(client, msg, notify)
	<-time.After(time.Duration(cacheTimer) * time.Second)
	deleteMsg(client, respMsg.FromChatID(), respMsg.Id)
}

func sendDisconnectMsg(client *http.Client, chat *models.Chat, ik *models.InlineKeyboardMarkup) {

	msg, err := models.NewMsg(
		models.WithMsgChat(chat),
		models.WithMsgText("Choose alerts to disconnect"),
		models.WithMsgReplyMarkup(ik),
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	sendMsg(client, msg, false)
}

func EditMarkup(client *http.Client, callback *models.CallbackQuery, pairs []string) error {
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

	editMarkup, err := models.NewEditMSGReplyMarkup(
		models.WithEMOChatID(callback.Msg.FromChatID()),
		models.WithEMOMsgID(callback.Msg.Id),
		models.WithEMOReplyMarkup(ik),
	)
	if err != nil {
		return err
	}

	editMSGReplyMarkup(client, editMarkup)
	return nil
}

func editMSGReplyMarkup(client *http.Client, editMarkup *models.EditMarkupObj) {
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

func composeKeyboardMarkup(pairs []string) (*models.InlineKeyboardMarkup, error) {
	if len(pairs) > 0 {
		inlineButtons := make([]models.InlineKeyboardButton, len(pairs)+1)

		disconnectAllBut, err := models.NewInlineKeyboardButton(
			models.WithIKBText("All"),
			models.WithIKBCallbackData("disconnect all"),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}

		inlineButtons[len(inlineButtons)-1] = *disconnectAllBut

		for i, v := range pairs {
			ikb, err := models.NewInlineKeyboardButton(
				models.WithIKBText(v),
				models.WithIKBCallbackData(fmt.Sprintf("disconnect %s", v)),
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

		inlineKeyboard := make([][]models.InlineKeyboardButton, size)

		for i := range inlineKeyboard {
			size = 3
			if len(inlineButtons) <= 3 {
				size = len(inlineButtons)
			}
			inlineKeyboard[i], inlineButtons = inlineButtons[:size], inlineButtons[size:]
		}

		ik, err := models.NewInlineKeyboardMarkup(inlineKeyboard)
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
	data, err := json.Marshal(models.NewDeleteMsgObj(chatID, msgID))
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

	ok := &models.DeleteResult{}

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
