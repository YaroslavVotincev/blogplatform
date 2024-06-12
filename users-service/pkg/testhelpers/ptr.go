package testhelpers

import "time"

func StrPtr(str string) *string {
	ptr := new(string)
	*ptr = str
	return ptr
}

func TimePtr(t time.Time) *time.Time {
	ptr := new(time.Time)
	*ptr = t
	return ptr
}
