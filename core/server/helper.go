package server

import (
	"github.com/Allenxuxu/gev"
	"go.uber.org/atomic"
)

const (
	Id = "ws:id"
)

var (
	ider atomic.Uint64
)

func setId(c *gev.Connection) (id uint64) {
	id = ider.Inc()
	c.Set(Id, id)
	return
}

func GetId(c *gev.Connection) (id uint64, exists bool) {
	var value interface{}
	value, exists = c.Get(Id)
	if !exists {
		return 0, exists
	}

	return value.(uint64), exists
}
