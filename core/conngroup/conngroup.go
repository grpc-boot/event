package conngroup

import "go.uber.org/atomic"

const (
	length = 9
	size   = 1 << length
	max    = size - 1
)

type ConnGroup struct {
	shards [size]shard
	length atomic.Int64
}

func NewConnGroup() *ConnGroup {
	cg := &ConnGroup{}
	for index := 0; index < size; index++ {
		cg.shards[0] = shard{}
	}

	return cg
}

func (cg *ConnGroup) index(id uint64) int {
	return int(id & max)
}

func (cg *ConnGroup) Exists(id uint64) bool {
	index := cg.index(id)
	return cg.shards[index].exists(id)
}

func (cg *ConnGroup) Set(id uint64, value interface{}) (isCreate bool) {
	index := cg.index(id)
	isCreate = cg.shards[index].set(id, value)
	if isCreate {
		cg.length.Inc()
	}
	return
}

func (cg *ConnGroup) Get(id uint64) (value interface{}, exists bool) {
	index := cg.index(id)
	return cg.shards[index].get(id)
}

func (cg *ConnGroup) Delete(id uint64) (delNum int) {
	index := cg.index(id)
	delNum = cg.shards[index].delete(id)
	if delNum > 0 {
		cg.length.Dec()
	}

	return
}

func (cg *ConnGroup) RangeValues(handler func(values []interface{})) {
	for index := 0; index < size; index++ {
		handler(cg.shards[index].values())
	}
}

func (cg *ConnGroup) Length() int64 {
	return cg.length.Load()
}
