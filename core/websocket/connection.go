package websocket

import (
	"net"
	"sync"
	"time"

	"event/core/base"
	"event/core/protocol"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	base2 "github.com/grpc-boot/base"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type Connection interface {
	Id() string
	RemoteIp() string
	RemotePort() uint16
	RemoteAddr() net.Addr
	Write(p []byte) (int, error)
	WriteJson(data interface{}) error
	WriteTimeout(p []byte, timeout time.Duration) (int, error)
	Receive() error
	Read(p []byte) (int, error)
	ReadJson(out interface{}) error
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
	conn       net.Conn
	mutex      sync.RWMutex
	ioMutex    sync.Mutex
	data       base.Params
	id         string
	remoteIp   string
	remotePort uint16
}

func (c *connection) Id() string {
	return c.id
}

func (c *connection) RemoteIp() string {
	return c.remoteIp
}

func (c *connection) RemotePort() uint16 {
	return c.remotePort
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
	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	return c.conn.Write(p)
}

func (c *connection) WriteJson(data interface{}) error {
	w := wsutil.NewWriter(c.conn, ws.StateServerSide, ws.OpText)
	encoder := jsoniter.NewEncoder(w)

	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return w.Flush()
}

func (c *connection) WriteTimeout(p []byte, timeout time.Duration) (int, error) {
	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	if err := c.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	return c.conn.Write(p)
}

func (c *connection) Receive() error {
	var req = &protocol.Request{}
	err := c.ReadJson(req)
	if err != nil {
		base2.ZapError("read json failed",
			zap.Error(err),
		)
		return err
	}

	if req == nil {
		return nil
	}

	base2.ZapInfo("receive msg",
		zap.ByteString("Msg", req.JsonMarshal()),
	)

	return nil
}

func (c *connection) Read(p []byte) (int, error) {
	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	return c.conn.Read(p)
}

func (c *connection) ReadJson(out interface{}) error {
	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	h, r, err := wsutil.NextReader(c.conn, ws.StateServerSide)
	if err != nil {
		return err
	}

	if h.OpCode.IsControl() {
		return wsutil.ControlFrameHandler(c.conn, ws.StateServerSide)(h, r)
	}

	decoder := jsoniter.NewDecoder(r)
	if err = decoder.Decode(out); err != nil {
		return err
	}

	return nil
}

func (c *connection) ReadTimeout(p []byte, timeout time.Duration) (int, error) {
	c.ioMutex.Lock()
	defer c.ioMutex.Unlock()

	if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return 0, err
	}
	return c.conn.Read(p)
}

func (c *connection) Reset() {
	c.mutex.Lock()
	c.data = make(base.Params)
	c.mutex.Unlock()

	c.id = ""
	c.remoteIp = ""
	c.remotePort = 0
	c.conn = nil
}

func (c *connection) Close() error {
	err := c.conn.Close()
	c.Reset()
	return err
}
