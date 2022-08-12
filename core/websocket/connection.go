package websocket

import (
	"net"
	"sync"
	"time"

	"event/core/base"
)

type Connection interface {
	RemoteAddr() net.Addr
	Write(p []byte) (int, error)
	WriteTimeout(p []byte, timeout time.Duration) (int, error)
	Read(p []byte) (int, error)
	ReadTimeout(p []byte, timeout time.Duration) (int, error)
	SetData(key string, value interface{})
	Exists(key string) bool
	GetData(key string) (value interface{}, exists bool)
	GetInt(key string) (value int)
	GetInt64(key string) (value int64)
	GetString(key string) (value string)
	Reset()
	Close() error
}

func newConnection() Connection {
	return &connection{
		data: make(base.Params),
	}
}

type connection struct {
	conn  net.Conn
	mutex sync.RWMutex
	data  base.Params
}

func (c *connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *connection) SetData(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
}

func (c *connection) GetData(key string) (value interface{}, exists bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists = c.data[key]
	return
}

func (c *connection) Exists(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.data.Exists(key)
}

func (c *connection) GetInt(key string) (value int) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.data.GetInt(key)
}

func (c *connection) GetInt64(key string) (value int64) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.data.GetInt64(key)
}

func (c *connection) GetString(key string) (value string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.data.GetString(key)
}

func (c *connection) Write(p []byte) (int, error) {
	return c.conn.Write(p)
}

func (c *connection) WriteTimeout(p []byte, timeout time.Duration) (int, error) {
	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	return c.conn.Write(p)
}

func (c *connection) Read(p []byte) (int, error) {
	return c.conn.Read(p)
}

func (c *connection) ReadTimeout(p []byte, timeout time.Duration) (int, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	return c.conn.Read(p)
}

func (c *connection) Reset() {
	c.mutex.Lock()
	c.data = make(base.Params)
	c.mutex.Unlock()
	c.conn = nil
}

func (c *connection) Close() error {
	err := c.conn.Close()
	c.Reset()
	return err
}
