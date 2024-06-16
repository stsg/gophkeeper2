package model

type TokenHolder struct {
	token string
}

func (s *TokenHolder) Set(token string) {
	s.token = token
}

func (s *TokenHolder) Get() string {
	return s.token
}
