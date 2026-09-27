package usecase

import "concert-go/internal/repository"

type HelloUsecase interface {
	GetGreeting() (string, error)
}

type helloUsecaseImpl struct {
	helloRepo repository.HelloRepository
}

// NewHelloUsecase implements HelloUsecase
func NewHelloUsecase(helloRepo repository.HelloRepository) HelloUsecase {
	return &helloUsecaseImpl{
		helloRepo: helloRepo,
	}
}

func (s *helloUsecaseImpl) GetGreeting() (string, error) {
	err := s.helloRepo.PingDatabase()
	if err != nil {
		return "", err
	}
	return "Hello from konserGo!", nil
}
