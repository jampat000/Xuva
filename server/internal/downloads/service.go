package downloads

type Job struct {
	ID            string `json:"id"`
	MediaSourceID string `json:"mediaSourceId"`
	TargetProfile string `json:"targetProfile"`
	Status        string `json:"status"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
