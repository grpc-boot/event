package conngroup

import "sync"

type shard struct {
	mutex sync.RWMutex
	items map[interface{}]interface{}
}

func (s *shard) exists(key interface{}) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, exists := s.items[key]
	return exists
}

func (s *shard) get(key interface{}) (value interface{}, exists bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists = s.items[key]
	return
}

func (s *shard) set(key interface{}, value interface{}) (isCreate bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.items == nil {
		s.items = map[interface{}]interface{}{}
		s.items[key] = value
		return true
	}

	_, exists := s.items[key]
	s.items[key] = value

	return !exists
}

func (s *shard) delete(keyList ...interface{}) (delNum int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, key := range keyList {
		if _, exists := s.items[key]; exists {
			delNum++
			delete(s.items, key)
		}
	}

	return
}

func (s *shard) values() (items []interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	items = make([]interface{}, len(s.items), len(s.items))
	if len(s.items) < 1 {
		return
	}

	index := 0
	for key, _ := range s.items {
		items[index] = s.items[key]
		index++
	}

	return
}
