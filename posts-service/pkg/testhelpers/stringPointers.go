package testhelpers

func StrPtr(str string) *string {
	ptr := new(string)
	*ptr = str
	return ptr
}