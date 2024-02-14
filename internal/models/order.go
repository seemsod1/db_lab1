package models

type Order struct {
	Country [10]byte
	Next    int64
	UserId  uint32
	RentId  uint32
	Price   uint32
	Deleted bool
}
