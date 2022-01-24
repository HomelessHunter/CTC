package models

import "errors"

type CallbackAnswer struct {
	Id        string `json:"callback_query_id"`
	Text      string `json:"text,omitempty"`
	ShowAlert bool   `json:"show_alert,omitempty"`
	URL       string `json:"url,omitempty"`
	CacheTime int    `json:"cache_time,omitempty"`
}

func NewCallbackAnswer(opts ...CallbackAnswerOpts) (*CallbackAnswer, error) {
	answer := CallbackAnswer{}

	for _, opt := range opts {
		err := opt(&answer)
		if err != nil {
			return nil, err
		}
	}

	return &answer, nil
}

type CallbackAnswerOpts func(*CallbackAnswer) error

func WithAnswerID(id string) CallbackAnswerOpts {
	return func(ca *CallbackAnswer) error {
		if id == "" {
			return errors.New("id shouldn't be empty")
		}

		ca.Id = id
		return nil
	}
}

func WithAnswerText(text string) CallbackAnswerOpts {
	return func(ca *CallbackAnswer) error {
		if text == "" {
			return errors.New("text shouldn't be empty")
		}

		ca.Text = text
		return nil
	}
}

func WithAnswerShowAlert(showAlert bool) CallbackAnswerOpts {
	return func(ca *CallbackAnswer) error {
		ca.ShowAlert = showAlert
		return nil
	}
}

func WithAnswerURL(url string) CallbackAnswerOpts {
	return func(ca *CallbackAnswer) error {
		if url == "" {
			return errors.New("url shouldn't be empty")
		}

		ca.URL = url
		return nil
	}
}

func WithAnswerCacheTime(cacheTime int) CallbackAnswerOpts {
	return func(ca *CallbackAnswer) error {
		if cacheTime < 0 {
			return errors.New("cacheTime should be positive")
		}

		ca.CacheTime = cacheTime
		return nil
	}
}
