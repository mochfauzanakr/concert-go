package service

import "konserGo/internal/repository"

type HelloService interface {
	GetGreeting() (string, error)
}

type helloServiceImpl struct {
	helloRepo repository.HelloRepository
}

// NewHelloService implements HelloService
func NewHelloService(helloRepo repository.HelloRepository) HelloService {
	return &helloServiceImpl{
		helloRepo: helloRepo,
	}
}

func (s *helloServiceImpl) GetGreeting() (string, error) {
	err := s.helloRepo.PingDatabase()
	if err != nil {
		return "", err
	}
	return "Hello from konserGo!", nil
}
