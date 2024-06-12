package configservice

import "time"

type ServiceModel struct {
	Service string    `json:"service"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Setting struct {
	Key     string    `json:"key"`
	Value   string    `json:"value"`
	Service string    `json:"-"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
