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

	db "github.com/HomelessHunter/CTC/db/models"
	other "github.com/HomelessHunter/CTC/wrapper/models/other"
	telegram "github.com/HomelessHunter/CTC/wrapper/models/telegram"
)

var ErrEmptyPairs = errors.New("pairs shouldn't be empty")

func SetWebhook(client *http.Client) {
	apiKey := os.Getenv("TG")
	postBody, err := json.Marshal(map[string]string{
		"url": fmt.Sprintf("https://main.myserversuck.keenetic.link/%s", apiKey),
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

func CommandRouter(command string, regs map[string]*regexp.Regexp) string {
	switch {
	case regs["start"].MatchString(command):
		return "start"

	case regs["help"].MatchString(command):
		return "help"

	case regs["alert"].MatchString(command):
		return "alert"

	case regs["price"].MatchString(command):
		return "price"

	case regs["disconnect"].MatchString(command):
		return "disconnect"
	}
	return ""
}

func StartRouter(update *telegram.Update, client *http.Client) error {
	msg, err := telegram.NewMsg(telegram.WithMsgText("Hello! I'm your personal crypto companion\nThat's what i can do for you:\n"), telegram.WithMsgChat(update.FromChat()))
	if err != nil {
		return fmt.Errorf("StartRouter: %v", err)
	}

	_, err = sendMsg(client, *msg, false)
	if err != nil {
		return fmt.Errorf("StartRouter: %v", err)
	}
	err = HelpRouter(update, client)
	if err != nil {
		return fmt.Errorf("StartRouter: %v", err)
	}
	return nil
}

func HelpRouter(update *telegram.Update, client *http.Client) error {
	text := "&#128142; <b>CryptoTrader Companion</b> &#128142;\n\n&#128073; <b>ALERT</b>\nType <b><u>/alert &#60;pair/symbols&#62 &#60target price&#62</u></b> to set alert (e.g. <u>/alert btcusdt 53400</u>)\n\n&#128073; <b>PRICE</b>\nType <b><u>/price &#60pair/symbols&#62</u></b> (e.g. <u>/price ehtbusd</u>)"
	msg, err := telegram.NewMsg(telegram.WithMsgText(text), telegram.WithMsgChat(update.FromChat()))
	if err != nil {
		return fmt.Errorf("HelpRouter: %s", err)
	}
	_, err = sendMsg(client, *msg, false)
	return err
}

func AlertRouter(command string, regs map[string]*regexp.Regexp, update *telegram.Update, client *http.Client) (*other.WSQuery, error) {
	c := regs["splitter"].Split(command, 3)
	price, err := strconv.ParseFloat(c[2], 64)
	if err != nil {
		return nil, fmt.Errorf("AlertRouter: %s", err)
	}

	market, err := getMarket(c[1], client)
	if err != nil {
		sendNoPairErr(client, *update.FromChat(), c[1])
		return nil, fmt.Errorf("AlertRouter: %s", err)
	}

	wsQuery, err := other.NewWsQuery(
		other.WithWSUserId(update.FromUser().ID()),
		other.WithWSChatId(update.FromChat().ID()),
		other.WithWSMarket(market),
		other.WithWSPair(c[1]),
		other.WithWSPrice(price),
	)
	if err != nil {
		return nil, fmt.Errorf("AlertRouter: %s", err)
	}
	return wsQuery, nil
}

func DisconnectRouter(update *telegram.Update, pairs []db.Alert, client *http.Client) error {
	ik, err := composeKeyboardMarkup(pairs)
	if err != nil {
		if err == ErrEmptyPairs {
			msg, err := telegram.NewMsg(
				telegram.WithMsgId(update.Id),
				telegram.WithMsgChat(update.FromChat()),
				telegram.WithMsgText("You have no alerts"),
			)
			if err != nil {
				return fmt.Errorf("DisconnectRouter: %s", err)
			}
			err = sendNDiscardMsg(client, *msg, false, 2)
			if err != nil {
				return fmt.Errorf("DisconnectRouter: %s", err)
			}
			return err
		}
		return fmt.Errorf("DisconnectRouter: %s", err)
	}
	err = sendDisconnectMsg(client, update.FromChat(), ik)
	if err != nil {
		return fmt.Errorf("DisconnectRouter: %s", err)
	}
	return nil
}

func PriceRouter(client *http.Client, update *telegram.Update, command string, regs map[string]*regexp.Regexp) error {
	symbol := regs["splitter"].Split(command, 2)[1]
	price, market, err := getLatestPrice(symbol, client)
	if err != nil {
		sendNoPairErr(client, *update.FromChat(), strings.ToUpper(symbol))
		return fmt.Errorf("PriceRouter: %s\n", err)
	}
	decal := "&#128310;"
	if market == Huobi {
		decal = "&#128309;"
	}
	msgText := fmt.Sprintf("%s <b>%s</b> - <b>%.2f</b>\n", decal, strings.ToUpper(symbol), price)
	if price < 1 {
		msgText = fmt.Sprintf("%s <b>%s</b> - <b>%f</b>\n", decal, strings.ToUpper(symbol), price)
	}
	msg, err := telegram.NewMsg(telegram.WithMsgText(msgText), telegram.WithMsgChat(update.FromChat()))
	if err != nil {
		return fmt.Errorf("PriceRouter: %s\n", err)
	}
	_, err = sendMsg(client, *msg, false)
	if err != nil {
		return fmt.Errorf("PriceRouter: %s\n", err)
	}
	return nil
}

func CallbackHandler(client *http.Client, callbackData string, regs map[string]*regexp.Regexp) string {
	switch {
	case regs["disconnect"].MatchString(callbackData):
		return "disconnect"
	default:
		return ""
	}
}

func DisconnectCallback(update *telegram.Update, regs map[string]*regexp.Regexp, client *http.Client) (*telegram.CallbackQuery, error) {
	if update.GetCallbackData() == "" {
		return nil, errors.New("callback shouldn't be empty")
	}

	callbackData := regs["splitter"].Split(update.GetCallbackData(), 3)

	if len(callbackData) == 2 {
		update.CallbackQuery.SetData(callbackData[1])
		return &update.CallbackQuery, nil
	}
	update.CallbackQuery.SetData(fmt.Sprintf("%s %s", callbackData[1], callbackData[2]))
	return &update.CallbackQuery, nil
}

func SplitCallbackData(data string) (pair string, market string) {
	callback := strings.Split(data, " ")
	pair = callback[0]
	if len(callback) > 1 {
		market = callback[1]
	}
	return
}

func SendCallbackAnswer(client *http.Client, callbackAnswer *telegram.CallbackAnswer) error {
	data, err := json.Marshal(callbackAnswer)
	if err != nil {
		return fmt.Errorf("SendCallbackAnswer: %s", err)
	}

	body := bytes.NewReader(data)
	_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", os.Getenv("TG")), "application/json", body)
	if err != nil {
		return fmt.Errorf("SendCallbackAnswer: %s", err)
	}
	return nil
}

func SendAlert(client *http.Client, chatID int64, symbol string, price float64) error {
	chat, err := telegram.NewChat(telegram.WithChatId(chatID))
	if err != nil {
		return fmt.Errorf("SendAlert: %s", err)
	}
	msg, err := telegram.NewMsg(telegram.WithMsgText(fmt.Sprintf("&#128680; <b>%s</b> - <b>%.2f</b>", strings.ToUpper(symbol), price)), telegram.WithMsgChat(chat))
	if err != nil {
		return fmt.Errorf("SendAlert: %s", err)
	}
	_, err = sendMsg(client, *msg, true)
	if err != nil {
		return fmt.Errorf("SendAlert: %v", err)
	}
	return nil
}

func SendAlertConfirmed(client *http.Client, chatID int64) error {
	chat, err := telegram.NewChat(telegram.WithChatId(chatID))
	if err != nil {
		return fmt.Errorf("SendAlertConfirmed: %v", err)
	}
	msg, err := telegram.NewMsg(telegram.WithMsgText("<b>Alert has been set</b>"), telegram.WithMsgChat(chat))
	if err != nil {
		return fmt.Errorf("SendAlertConfirmed: %v", err)
	}
	_, err = sendMsg(client, *msg, false)
	if err != nil {
		return fmt.Errorf("SendAlertConfirmed: %v", err)
	}
	return nil
}

func sendNoPairErr(client *http.Client, chat telegram.Chat, pair string) error {
	msg, err := telegram.NewMsg(telegram.WithMsgChat(&chat), telegram.WithMsgText(fmt.Sprintf("Can't find <b>%s</b> &#129301;", pair)))
	if err != nil {
		return fmt.Errorf("sendNoPairErr: %v", err)
	}
	_, err = sendMsg(client, *msg, false)
	return err
}

func SendAlertExist(client *http.Client, chatID int64, pair string) error {
	chat, err := telegram.NewChat(telegram.WithChatId(chatID))
	if err != nil {
		return fmt.Errorf("SendAlertExist: %v", err)
	}
	msg, err := telegram.NewMsg(telegram.WithMsgChat(chat), telegram.WithMsgText(fmt.Sprintf("&#9940; %s alert already exist", pair)))
	if err != nil {
		return fmt.Errorf("SendAlertExist: %v", err)
	}
	_, err = sendMsg(client, *msg, false)
	if err != nil {
		return fmt.Errorf("SendAlertExist: %v", err)
	}
	return nil
}

func sendMsg(client *http.Client, msg telegram.Message, notify bool) (*telegram.Message, error) {
	sendObj, err := telegram.NewSendMsgObj(
		telegram.WithSendChatId(msg.FromChatID()),
		telegram.WithSendText(msg.Text),
		telegram.WithSendParseMode("HTML"),
		telegram.WithSendDisableNotification(notify),
		telegram.WithSendAllowReply(true),
		telegram.WithSendReplyMarkup(msg.ReplyMarkup),
	)

	if err != nil {
		return nil, fmt.Errorf("sendMsg: %s", err)
	}

	data, err := json.Marshal(sendObj)
	if err != nil {
		return nil, fmt.Errorf("sendMsg: %s", err)
	}
	body := bytes.NewReader(data)
	resp, err := client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TG")), "application/json", body)
	if err != nil {
		return nil, fmt.Errorf("sendMsg: %s", err)
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sendMsg: %s", err)
	}
	defer resp.Body.Close()

	responseMsg := telegram.NewResponseMessage()

	err = json.Unmarshal(data, responseMsg)
	if err != nil {
		return nil, fmt.Errorf("sendMsg: %s", err)
	}

	return &responseMsg.Result, nil
}

func sendNDiscardMsg(client *http.Client, msg telegram.Message, notify bool, cacheTimer int) error {
	respMsg, err := sendMsg(client, msg, notify)
	if err != nil {
		return fmt.Errorf("sendNDiscardMsg: %s", err)
	}
	<-time.After(time.Duration(cacheTimer) * time.Second)
	deleteMsg(client, respMsg.FromChatID(), respMsg.Id)
	return nil
}

func sendDisconnectMsg(client *http.Client, chat *telegram.Chat, ik *telegram.InlineKeyboardMarkup) error {

	msg, err := telegram.NewMsg(
		telegram.WithMsgChat(chat),
		telegram.WithMsgText("Choose alerts to disconnect"),
		telegram.WithMsgReplyMarkup(ik),
	)

	if err != nil {
		return fmt.Errorf("sendDisconnectMsg: %s", err)
	}

	_, err = sendMsg(client, *msg, false)
	if err != nil {
		return fmt.Errorf("sendDisconnectMsg: %s", err)
	}

	return nil
}

func EditMarkup(client *http.Client, callback *telegram.CallbackQuery, pairs []db.Alert) error {
	if len(pairs) == 0 {
		// delete msg
		fmt.Println(len(pairs))
		if !deleteMsg(client, callback.Msg.FromChatID(), callback.Msg.Id) {
			return errors.New("could not delete a message")
		}
		return nil
	}
	ik, err := composeKeyboardMarkup(pairs)
	if err != nil {
		return fmt.Errorf("EditMarkup: %s", err)
	}

	editMarkup, err := telegram.NewEditMSGReplyMarkup(
		telegram.WithEMOChatID(callback.Msg.FromChatID()),
		telegram.WithEMOMsgID(callback.Msg.Id),
		telegram.WithEMOReplyMarkup(ik),
	)
	if err != nil {
		return fmt.Errorf("EditMarkup: %s", err)
	}

	editMSGReplyMarkup(client, editMarkup)
	fmt.Println("EDIT_MARKUP")
	return nil
}

func editMSGReplyMarkup(client *http.Client, editMarkup *telegram.EditMarkupObj) error {
	data, err := json.Marshal(editMarkup)
	if err != nil {
		return fmt.Errorf("editMSGReplyMarkup: %s", err)
	}

	body := bytes.NewReader(data)
	_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/editMessageReplyMarkup", os.Getenv("TG")), "application/json", body)
	if err != nil {
		return fmt.Errorf("editMSGReplyMarkup: %s", err)
	}
	return nil
}

func composeKeyboardMarkup(pairs []db.Alert) (*telegram.InlineKeyboardMarkup, error) {
	if len(pairs) > 0 {
		inlineButtons := make([]telegram.InlineKeyboardButton, len(pairs)+1)

		disconnectAllBut, err := telegram.NewInlineKeyboardButton(
			telegram.WithIKBText("All"),
			telegram.WithIKBCallbackData("disconnect all"),
		)
		if err != nil {
			return nil, fmt.Errorf("composeKeyboardMarkup: %s", err)
		}

		inlineButtons[len(inlineButtons)-1] = *disconnectAllBut

		for i, v := range pairs {
			ikb, err := telegram.NewInlineKeyboardButton(
				telegram.WithIKBText(v.Pair),
				telegram.WithIKBCallbackData(fmt.Sprintf("disconnect %s %s", v.Pair, v.Market)),
			)
			if err != nil {
				return nil, fmt.Errorf("composeKeyboardMarkup: %s", err)
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
			return nil, fmt.Errorf("composeKeyboardMarkup: %s", err)
		}

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
