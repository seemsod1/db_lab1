package models

type Order struct {
	Next   int64
	UserId uint32
	RentId uint32
	Price  uint32
}
