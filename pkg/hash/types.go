package hash

type Service interface {
	HashSlice(data interface{}) (string, error)
}
