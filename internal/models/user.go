package models

type User struct {
	FirstOrder int64
	ID         uint32
	Age        uint32
	Name       [20]byte
	Mail       [20]byte
	Deleted    bool
}
