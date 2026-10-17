package gateway

type CodeHasher interface {
	Hash(code string) ([]byte, error)
	Compare(hash []byte, code string) bool
}
