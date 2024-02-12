package models

type User struct {
	ID   uint32
	Name [20]byte
	Mail [20]byte
	Age  uint32

	FirstOrder int64
}
