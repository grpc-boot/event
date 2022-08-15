package server

import "github.com/Allenxuxu/gev"

const (
	Id = "ws:id"
)

func setId(c *gev.Connection, id int64) {
	c.Set(Id, id)
}

func getId(c *gev.Connection) (id int64, exists bool) {
	var value interface{}
	value, exists = c.Get(Id)
	if !exists {
		return 0, exists
	}

	return value.(int64), exists
}
