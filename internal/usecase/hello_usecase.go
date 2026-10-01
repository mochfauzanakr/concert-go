package usecase

import "concert-go/internal/repository"

type HelloUsecase struct {
	helloRepo repository.HelloRepository
}

// NewHelloUsecase implements HelloUsecase
func NewHelloUsecase(helloRepo repository.HelloRepository) *HelloUsecase {
	return &HelloUsecase{
		helloRepo: helloRepo,
	}
}

func (s *HelloUsecase) GetGreeting() (string, error) {
	err := s.helloRepo.PingDatabase()
	if err != nil {
		return "", err
	}
	return "Hello from konserGo!", nil
}
