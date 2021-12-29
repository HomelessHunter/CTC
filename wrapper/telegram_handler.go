package wrapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

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
	update Update, rw http.ResponseWriter, client *http.Client, pairs []string) {

	switch {
	case regs["alert"].MatchString(command):
		c := regs["splitter"].Split(command, 3)
		price, err := strconv.ParseFloat(c[2], 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		wsQuery, err := json.Marshal(WSQuery{
			UserId: update.Msg.From.Id,
			ChatId: update.Msg.Chat.Id,
			Pair:   c[1],
			Price:  price,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		req, err := createRequest(wsQuery, "POST", "/connect")
		if err != nil {
			return
		}

		http.Redirect(rw, req, "/connect", http.StatusFound)

	case regs["disconnect"].MatchString(command):
		if len(pairs) > 0 {
			sendDisconnectMsg(client, update.Msg.SenderChat, pairs)
		} else {
			// sendMsg()
		}
	}
}

func CallbackHandler(data string, msg Message, regs map[string]*regexp.Regexp) {
	switch {
	case regs["disconnect"].MatchString(data):
		data = regs["splitter"].Split(data, 2)[1]
		// redirect to Disconnector
	}
}

func createRequest(data []byte, method string, url string) (*http.Request, error) {
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(method, url, body)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return req, nil
}

func sendMsg(client *http.Client, msg *Message, notify bool) {
	sendObj := &SendMsgObj{
		ChatId:                msg.Chat.Id,
		Text:                  msg.Text,
		ParseMode:             "HTML",
		Entities:              msg.Entities,
		DisableNotification:   notify,
		AllowSendWithoutReply: true,
		ReplyMarkup:           msg.ReplyMarkup,
	}

	data, err := json.Marshal(sendObj)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg: ", err)
		return
	}
	body := bytes.NewReader(data)
	_, err = client.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", os.Getenv("TG")), "application/json", body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "SendMsg: ", err)
	}
}

func sendDisconnectMsg(client *http.Client, chat Chat, pairs []string) {

	inlineButtons := make([]InlineKeyboardButton, len(pairs)+1)
	inlineButtons[len(inlineButtons)-1] = InlineKeyboardButton{
		Text:         "All",
		CallbackData: "disconnect all",
	}

	for i, v := range pairs {
		inlineButtons[i] = InlineKeyboardButton{
			Text:         v,
			CallbackData: fmt.Sprintf("disconnect %s", v),
		}
	}

	size := 3

	switch {
	case (len(pairs) % size) == 0:
		size = len(pairs) / size
	default:
		size = len(pairs)/size + 1
	}

	inlineKeyboard := make([][]InlineKeyboardButton, size)

	for i := range inlineKeyboard {
		size = 3
		if len(inlineButtons) <= 3 {
			size = len(inlineButtons)
		}
		inlineKeyboard[i], inlineButtons = inlineButtons[:size], inlineButtons[size:]
	}

	msg := &Message{
		Chat:        chat,
		Text:        "Choose alerts to disconnect",
		ReplyMarkup: InlineKeyboardMarkup{inlineKeyboard},
	}

	sendMsg(client, msg, false)
}
