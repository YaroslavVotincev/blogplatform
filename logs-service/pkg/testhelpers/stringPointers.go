package testhelpers

import "github.com/google/uuid"

func StrPtr(str string) *string {
	ptr := new(string)
	*ptr = str
	return ptr
}

func NewUUIDPtr() *uuid.UUID {
	ptr := new(uuid.UUID)
	*ptr = uuid.New()
	return ptr
}
