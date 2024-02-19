package models

type SHeader struct {
	Prev int64
	Pos  int64
	Next int64
}

type Order struct {
	OrderId uint32
	UserId  uint32
	Price   uint32
	Country [10]byte
	Next    int64
	Deleted bool
}

type User struct {
	ID         uint32
	Age        uint32
	Name       [10]byte
	Mail       [30]byte
	FirstOrder int64
	Deleted    bool
}
