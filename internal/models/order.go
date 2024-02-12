package models

type Order struct {
	UserId uint32
	Rent   uint32
	Price  uint32

	Next int64
}
