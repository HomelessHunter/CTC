package models

import (
	"context"
	"errors"
	"sync"
)

var ErrEmptyMap = errors.New("userChannels map shouldn't be empty")

type UC struct {
	mu           sync.RWMutex
	userChannels map[int64]*UserChannels
}

func NewUC() *UC {
	return &UC{userChannels: make(map[int64]*UserChannels)}
}

func (uc *UC) SetUC(id int64, userChannels *UserChannels) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.userChannels[id] = userChannels
}

func (uc *UC) GetUserChannels(id int64) (*UserChannels, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	userChannels, ok := uc.userChannels[id]
	return userChannels, ok
}

type UserChannels struct {
	cancel map[string]context.CancelFunc
	// ShutdownCh is giving a signal to close current websocket connection completely
	shutdownCh map[string]chan int
	// subscribeCh is giving a signal about subscribing a new pair
	subscribeCh map[string]chan PairSignal
	// unsubscribeCh is givin a signal about unsubscribing a pair
	unsubscribeCh map[string]chan PairSignal
}

type userChannels struct {
	Cancel        map[string]context.CancelFunc
	ShutdownCh    map[string]chan int
	SubscribeCh   map[string]chan PairSignal
	UnsubscribeCh map[string]chan PairSignal
}

func NewUserChannels(opts ...UserChannelsOpts) (*UserChannels, error) {
	userChannels := userChannels{}

	for _, opt := range opts {
		err := opt(&userChannels)
		if err != nil {
			return nil, err
		}
	}

	return &UserChannels{
		cancel:        userChannels.Cancel,
		shutdownCh:    userChannels.ShutdownCh,
		subscribeCh:   userChannels.SubscribeCh,
		unsubscribeCh: userChannels.UnsubscribeCh,
	}, nil
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

func (userChannels *UserChannels) DeleteShutdownCh(market string) {
	delete(userChannels.shutdownCh, market)
}

func (userChannels *UserChannels) Shutdown(market string) {
	userChannels.shutdownCh[market] <- 1
}

func (userChannels *UserChannels) SubscribeCh(market string) chan PairSignal {
	return userChannels.subscribeCh[market]
}

func (userChannels *UserChannels) SetSubscriberCh(market string, subscribeCh chan PairSignal) {
	userChannels.subscribeCh[market] = subscribeCh
}

func (userChannels *UserChannels) DeleteSubscribeCh(market string) {
	delete(userChannels.subscribeCh, market)
}

func (userChannels *UserChannels) SubscribeSignal(market string, pair PairSignal) {
	userChannels.subscribeCh[market] <- pair
}

func (userChannels *UserChannels) UnsubscribeCh(market string) chan PairSignal {
	return userChannels.unsubscribeCh[market]
}

func (userChannels *UserChannels) SetUnsubscriberCh(market string, unsubscribeCh chan PairSignal) {
	userChannels.unsubscribeCh[market] = unsubscribeCh
}

func (userChannels *UserChannels) DeleteUnsubscribeCh(market string) {
	delete(userChannels.unsubscribeCh, market)
}

func (userChannels *UserChannels) UnsubscribeSignal(market string, pair PairSignal) {
	userChannels.unsubscribeCh[market] <- pair
}

type PairSignal struct {
	Pair string
	Size int
}

type UserChannelsOpts func(*userChannels) error

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

func WithUCSubscribeCh(subscribeCh map[string]chan PairSignal) UserChannelsOpts {
	return func(uc *userChannels) error {
		uc.SubscribeCh = subscribeCh
		return nil
	}
}

func WithUCUnsubscribeCh(unsubscribeCh map[string]chan PairSignal) UserChannelsOpts {
	return func(uc *userChannels) error {
		uc.UnsubscribeCh = unsubscribeCh
		return nil
	}
}
