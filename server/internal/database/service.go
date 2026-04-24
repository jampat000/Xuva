package database

type Service struct {
	DataDir string
}

func NewService(dataDir string) *Service {
	return &Service{DataDir: dataDir}
}
