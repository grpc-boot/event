package protocol

import (
	"github.com/grpc-boot/base"
)

type Request struct {
	Id     string         `json:"id"`
	Name   string         `json:"name"`
	Params base.JsonParam `json:"params"`
}

func (r *Request) JsonMarshal() []byte {
	data, _ := base.JsonMarshal(r)
	return data
}
