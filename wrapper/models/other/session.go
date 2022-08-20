package models

import (
	"errors"
	"fmt"
	"sync"

	dbModels "github.com/HomelessHunter/CTC/db/models"
)

type Session struct {
	mu          sync.RWMutex
	alerts      map[int64]*[]dbModels.Alert
	alertsCount int
	markets     map[int64]*[]market
}

type market struct {
	market string
	exist  bool
}

func NewSession() *Session {
	return &Session{alerts: make(map[int64]*[]dbModels.Alert), alertsCount: 0, markets: make(map[int64]*[]market)}
}

func (s *Session) Alerts() map[int64]*[]dbModels.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alerts
}

func (s *Session) AlertsByID(id int64) []dbModels.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alerts := s.alerts[id]
	if alerts == nil {
		return nil
	}
	return *alerts
}

func (s *Session) AlertsByMarket(id int64, market string) (alerts []dbModels.Alert, newAlerts []dbModels.Alert) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	alertsP, ok := s.alerts[id]
	if !ok {
		newAlerts = nil
		return
	}
	alerts = *alertsP
	if len(alerts) == 0 {
		newAlerts = nil
		return
	}
	oldAlerts := make([]dbModels.Alert, len(alerts))
	lenght := 0
	for _, v := range alerts {
		if v.Market == market {
			oldAlerts[lenght] = v
			lenght++
		}
	}
	newAlerts = make([]dbModels.Alert, lenght)
	copy(newAlerts, oldAlerts)
	return
}

func (s *Session) AlertsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alertsCount
}

func (s *Session) InitMarketsByID(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println(id)
	if s.markets[id] == nil {
		s.markets[id] = &[]market{
			{market: "binance", exist: false},
			{market: "huobi", exist: false},
		}
	}
}

func (s *Session) SetMarketByID(id int64, market string, exist bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	markets := *s.markets[id]
	switch market {
	case "binance":
		markets[0].exist = exist
	case "huobi":
		markets[1].exist = exist
	}
}

func (s *Session) SetMarketsByID(id int64, exist bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	markets := *s.markets[id]
	for i := range markets {
		markets[i].exist = exist
	}
}

func (s *Session) MarketExist(id int64, market string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	markets := *s.markets[id]
	switch market {
	case "binance":
		return markets[0].exist
	case "huobi":
		return markets[1].exist
	}
	return false
}

func (s *Session) MarketsByID(id int64) ([]market, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	markets := *s.markets[id]
	if markets[0].market == "" {
		return nil, errors.New("")
	}
	return markets, nil
}

// func (s *Session) DeleteMarkets(id int64) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()
// 	delete(s.markets, id)
// }

func (s *Session) AddAlerts(id int64, alerts ...dbModels.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.alerts[id] == nil {
		sessionAlerts := make([]dbModels.Alert, 0, 1)
		s.alerts[id] = &sessionAlerts
	}
	for i := range alerts {
		alerts[i].Connected = true
	}
	*s.alerts[id] = append(*s.alerts[id], alerts...)
	s.alertsCount += len(alerts)
}

func (s *Session) DeleteAlert(id int64, index int) error {
	s.mu.RLock()
	alerts := *s.alerts[id]
	// dbModels.SortAlerts(alerts)
	fmt.Println("DeleteAlert", alerts)
	fmt.Println("DeleteAlert", alerts[index])
	if index >= len(alerts) {
		s.mu.RUnlock()
		return fmt.Errorf("%d index out of range", index)
	}
	s.mu.RUnlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	oldAlerts := append(alerts[:index], alerts[index+1:]...)
	newAlerts := make([]dbModels.Alert, len(oldAlerts))
	copy(newAlerts, oldAlerts)
	fmt.Println("NEW ALERTS", newAlerts)
	*s.alerts[id] = newAlerts
	fmt.Println("ALERTTS", *s.alerts[id])
	s.alertsCount -= 1
	return nil
}

func (s *Session) DeleteAlerts(id int64, length int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println(s.alerts)
	delete(s.alerts, id)
	if s.alertsCount > 0 {
		s.alertsCount -= length
	}
}
