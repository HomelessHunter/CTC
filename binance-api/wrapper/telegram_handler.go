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
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer resp.Body.Close()
	fmt.Fprintln(os.Stdout, string(data))
}

func CommandRouter(command string, regs map[string]*regexp.Regexp, update Update, rw http.ResponseWriter, r *http.Request) {

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

		body := bytes.NewBuffer(wsQuery)
		req, err := http.NewRequest("POST", "/connect", body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		http.Redirect(rw, req, "/connect", http.StatusFound)

	case regs["disconnect"].MatchString(command):
		data, err := json.Marshal(update)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		body := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", "/disconnect", body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		http.Redirect(rw, req, "/disconnect", http.StatusFound)
	}
}
