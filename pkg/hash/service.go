package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) HashSlice(data interface{}) (string, error) {
	jsonData, _ := json.Marshal(data)

	hash := sha256.Sum256(jsonData)

	return fmt.Sprintf("%x", hash), nil
}
