package sessions

import "time"

type Session struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	DeviceID      string    `json:"deviceId"`
	MediaSourceID string    `json:"mediaSourceId"`
	Mode          string    `json:"mode"`
	StartedAt     time.Time `json:"startedAt"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
