package libraries

type Library struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}
