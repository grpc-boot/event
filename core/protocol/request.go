package protocol

import (
	"event/core/base"

	jsoniter "github.com/json-iterator/go"
)

type Request struct {
	Id     string      `json:"id"`
	Name   string      `json:"name"`
	Params base.Params `json:"params"`
}

func (r *Request) JsonMarshal() []byte {
	data, _ := jsoniter.Marshal(r)
	return data
}
