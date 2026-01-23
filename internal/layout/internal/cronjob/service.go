package cronjob

type Service struct{}

func (s *Service) Run() error {
	return nil
}

func NewService() *Service {
	return &Service{}
}
