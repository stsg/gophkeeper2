package enum

type ResourceType uint8

const (
	Nan ResourceType = iota
	LoginPassword
	File
	BankCard
)
