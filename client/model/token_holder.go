// package model contains model types
package model

// The TokenHolder struct is a simple container that holds a string token.
type TokenHolder struct {
	token string
}

// Set sets the token for the TokenHolder.
//
// Parameter:
//
//	token string - the token to set.
func (s *TokenHolder) Set(token string) {
	s.token = token
}

// Get returns the token stored in the TokenHolder struct.
//
// No parameters.
// Returns a string representing the token.
func (s *TokenHolder) Get() string {
	return s.token
}
