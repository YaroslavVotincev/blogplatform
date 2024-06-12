package testhelpers

import (
	"encoding/json"
	"io"
	"strings"
)

func ReqBody(obj any) io.Reader {
	reqBodyBytes, _ := json.Marshal(obj)
	return strings.NewReader(string(reqBodyBytes))
}
