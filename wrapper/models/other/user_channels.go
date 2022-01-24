package models

import (
	"context"
	"errors"
)

type UserChannels struct {
	chatId int64
	pairs  []string
	cancel map[string]context.CancelFunc
	// ShutdownCh is giving a signal to close current websocket connection completely
	shutdownCh map[string]chan int
	// ReconnectCh is giving a signal to close current connection and reconnect with new values
	reconnectCh map[string]chan int
	// addPairCh is giving a signal about pair appending
	addPairCh map[string]chan int
}

type userChannels struct {
	ChatId      int64
	Pairs       []string
	Cancel      map[string]context.CancelFunc
	ShutdownCh  map[string]chan int
	ReconnectCh map[string]chan int
	addPairCh   map[string]chan int
}

func NewuserChannels(opts ...UserChannelsOpts) (*UserChannels, error) {
	userChannels := userChannels{}

	for _, opt := range opts {
		err := opt(&userChannels)
		if err != nil {
			return nil, err
		}
	}

	return &UserChannels{chatId: userChannels.ChatId,
		pairs:       userChannels.Pairs,
		cancel:      userChannels.Cancel,
		shutdownCh:  userChannels.ShutdownCh,
		reconnectCh: userChannels.ReconnectCh}, nil
}

func (userChannels *UserChannels) ChatID() int64 {
	return userChannels.chatId
}

func (userChannels *UserChannels) SetChatID(chatId int64) {
	userChannels.chatId = chatId
}

func (userChannels *UserChannels) Pairs() []string {
	return userChannels.pairs
}

func (userChannels *UserChannels) SetPairs(pairs []string) {
	userChannels.pairs = pairs
}

func (userChannels *UserChannels) AddPairs(pair ...string) []string {
	return append(userChannels.pairs, pair...)
}

func (userChannels *UserChannels) GetCancel() map[string]context.CancelFunc {
	return userChannels.cancel
}

func (userChannels *UserChannels) SetCancel(market string, cancel context.CancelFunc) {
	userChannels.cancel[market] = cancel
}

func (userChannels *UserChannels) Cancel(market string) {
	userChannels.cancel[market]()
}

func (userChannels *UserChannels) ShutdownCh(market string) chan int {
	return userChannels.shutdownCh[market]
}

func (userChannels *UserChannels) SetShutdownCh(market string, ch chan int) {
	userChannels.shutdownCh[market] = ch
}

func (userChannels *UserChannels) Shutdown(market string) {
	userChannels.shutdownCh[market] <- 1
}

func (userChannels *UserChannels) ReconnectCh(market string) chan int {
	return userChannels.reconnectCh[market]
}

func (userChannels *UserChannels) SetReconnectCh(market string, ch chan int) {
	userChannels.reconnectCh[market] = ch
}

func (userChannels *UserChannels) Reconnect(market string) {
	userChannels.reconnectCh[market] <- 1
}

func (userChannels *UserChannels) AddPairCh(market string) chan int {
	return userChannels.addPairCh[market]
}

func (userChannels *UserChannels) SetAddPairCh(market string, addPairCh chan int) {
	userChannels.addPairCh[market] = addPairCh
}

func (userChannels *UserChannels) AddPairSignal(market string) {
	userChannels.addPairCh[market] <- 1
}

type UserChannelsOpts func(*userChannels) error

func WithUCChatId(chatId int64) UserChannelsOpts {
	return func(uc *userChannels) error {
		if chatId < 0 {
			return errors.New("chatId should be positive")
		}

		uc.ChatId = chatId
		return nil
	}
}

func WithUCPairs(pairs []string) UserChannelsOpts {
	return func(uc *userChannels) error {
		if len(pairs) == 0 {
			return errors.New("pairs shouldn't be empty")
		}

		uc.Pairs = pairs
		return nil
	}
}

func WithUCCancel(cancel map[string]context.CancelFunc) UserChannelsOpts {
	return func(uc *userChannels) error {
		if cancel == nil {
			return errors.New("cancel shouldn't be empty")
		}

		uc.Cancel = cancel
		return nil
	}
}

func WithUCShutdown(shutdownCh map[string]chan int) UserChannelsOpts {
	return func(uc *userChannels) error {
		uc.ShutdownCh = shutdownCh
		return nil
	}
}

func WithUCReconnect(reconnectCh map[string]chan int) UserChannelsOpts {
	return func(uc *userChannels) error {
		uc.ReconnectCh = reconnectCh
		return nil
	}
}

func WithUCAddPairCh(addPairCh map[string]chan int) UserChannelsOpts {
	return func(uc *userChannels) error {
		uc.addPairCh = addPairCh
		return nil
	}
}
