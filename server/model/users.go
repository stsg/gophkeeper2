package model

type User struct {
	Id       int32  `db:"id"`
	Username string `db:"username"`
	Password []byte `db:"password"`
}
