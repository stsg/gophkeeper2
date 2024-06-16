package consts

var UserIDCtxKey = &contextKey{"userID"}

type contextKey struct {
	name string
}
