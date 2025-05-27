package hash

type MockService struct {
	Hash string
	Err  error
}

func (s *MockService) HashSlice(data interface{}) (string, error) {
	return s.Hash, s.Err
}
