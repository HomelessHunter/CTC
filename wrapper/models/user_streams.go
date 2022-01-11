package models

import (
	"context"
	"errors"
)

type UserStreams struct {
	chatId int64
	pairs  []string
	cancel context.CancelFunc
	// ShutdownCh is giving a signal to close current websocket connection completely
	shutdownCh chan int
	// ReconnectCh is giving a signal to close current connection and reconnect with new values
	reconnectCh chan int
	// addPairCh is giving a signal about pair appending
	addPairCh chan int
}

type userStreams struct {
	ChatId      int64
	Pairs       []string
	Cancel      context.CancelFunc
	ShutdownCh  chan int
	ReconnectCh chan int
	addPairCh   chan int
}

func NewUserStreams(opts ...UserStreamsOpts) (*UserStreams, error) {
	userStreams := userStreams{}

	for _, opt := range opts {
		err := opt(&userStreams)
		if err != nil {
			return nil, err
		}
	}

	return &UserStreams{chatId: userStreams.ChatId,
		pairs:       userStreams.Pairs,
		cancel:      userStreams.Cancel,
		shutdownCh:  userStreams.ShutdownCh,
		reconnectCh: userStreams.ReconnectCh}, nil
}

func (userStreams *UserStreams) ChatID() int64 {
	return userStreams.chatId
}

func (userStreams *UserStreams) SetChatID(chatId int64) {
	userStreams.chatId = chatId
}

func (userStreams *UserStreams) Pairs() []string {
	return userStreams.pairs
}

func (userStreams *UserStreams) SetPairs(pairs []string) {
	userStreams.pairs = pairs
}

func (userStreams *UserStreams) AddPairs(pair ...string) []string {
	return append(userStreams.pairs, pair...)
}

func (userStreams *UserStreams) GetCancel() context.CancelFunc {
	return userStreams.cancel
}

func (userStreams *UserStreams) SetCancel(cancel context.CancelFunc) {
	userStreams.cancel = cancel
}

func (userStreams *UserStreams) Cancel() {
	userStreams.cancel()
}

func (userStream *UserStreams) ShutdownCh() chan int {
	return userStream.shutdownCh
}

func (userStream *UserStreams) SetShutdownCh(ch chan int) {
	userStream.shutdownCh = ch
}

func (userStreams *UserStreams) Shutdown() {
	userStreams.shutdownCh <- 1
}

func (userStreams *UserStreams) ReconnectCh() chan int {
	return userStreams.reconnectCh
}

func (userStreams *UserStreams) SetReconnectCh(ch chan int) {
	userStreams.reconnectCh = ch
}

func (userStreams *UserStreams) Reconnect() {
	userStreams.reconnectCh <- 1
}

func (userStreams *UserStreams) AddPairCh() chan int {
	return userStreams.addPairCh
}

func (userStreams *UserStreams) SetAddPairCh(addPairCh chan int) {
	userStreams.addPairCh = addPairCh
}

func (userStreams *UserStreams) AddPairSignal() {
	userStreams.addPairCh <- 1
}

type UserStreamsOpts func(*userStreams) error

func WithUSChatId(chatId int64) UserStreamsOpts {
	return func(us *userStreams) error {
		if chatId < 0 {
			return errors.New("chatId should be positive")
		}

		us.ChatId = chatId
		return nil
	}
}

func WithUSPairs(pairs []string) UserStreamsOpts {
	return func(us *userStreams) error {
		if len(pairs) == 0 {
			return errors.New("pairs shouldn't be empty")
		}

		us.Pairs = pairs
		return nil
	}
}

func WithUSCancel(cancel context.CancelFunc) UserStreamsOpts {
	return func(us *userStreams) error {
		if cancel == nil {
			return errors.New("cancel shouldn't be empty")
		}

		us.Cancel = cancel
		return nil
	}
}

func WithUSShutdown(shutdownCh chan int) UserStreamsOpts {
	return func(us *userStreams) error {
		us.ShutdownCh = shutdownCh
		return nil
	}
}

func WithUSReconnect(reconnectCh chan int) UserStreamsOpts {
	return func(us *userStreams) error {
		us.ReconnectCh = reconnectCh
		return nil
	}
}

func WithUSAddPairCh(addPairCh chan int) UserStreamsOpts {
	return func(us *userStreams) error {
		us.addPairCh = addPairCh
		return nil
	}
}
