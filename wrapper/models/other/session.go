package models

import (
	"fmt"
	"sync"

	dbModels "github.com/HomelessHunter/CTC/db/models"
)

type Session struct {
	mu          sync.Mutex
	alerts      map[int64]*[]dbModels.Alert
	alertsCount int
}

func NewSession() *Session {
	return &Session{alerts: make(map[int64]*[]dbModels.Alert), alertsCount: 0}
}

func (s *Session) Alerts(id int64) []dbModels.Alert {
	return *s.alerts[id]
}

func (s *Session) AddAlerts(id int64, alerts ...dbModels.Alert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.alerts[id] == nil {
		alerts := make([]dbModels.Alert, 0, 1)
		s.alerts[id] = &alerts
	}
	*s.alerts[id] = append(*s.alerts[id], alerts...)
	s.alertsCount += len(alerts)
}

func (s *Session) DeleteAlert(id int64, index int) error {
	alerts := *s.alerts[id]
	if index >= len(alerts) {
		return fmt.Errorf("%d index out of range", index)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	oldAlerts := append(alerts[:index], alerts[index+1:]...)
	newAlerts := make([]dbModels.Alert, len(oldAlerts))
	copy(newAlerts, oldAlerts)
	*s.alerts[id] = newAlerts
	s.alertsCount -= 1
	return nil
}

func (s *Session) DeleteAlerts(id int64, length int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.alerts, id)
	s.alertsCount -= length
}
