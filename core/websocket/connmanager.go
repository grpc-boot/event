package websocket

import (
	"net"
	"sync"
)

// ConnManager 连接池
type ConnManager interface {
	AcquireConn(c net.Conn) Connection
	ReleaseConn(conn Connection)
	Login(guestId string, userId int64, conn Connection)
	LoginOut(userId int64)
	GetGuest(guestId string) Connection
	GetUser(userId int64) Connection
	GuestTotal() int
	UserTotal() int
	ConnTotal() int
}

// NewConnManager 实例化连接池
func NewConnManager() ConnManager {
	return &connManager{
		pool: sync.Pool{
			New: func() interface{} {
				return newConnection()
			},
		},
	}
}

type connManager struct {
	mutex      sync.RWMutex
	guestConns map[string]Connection
	loginConns map[int64]Connection
	pool       sync.Pool
}

func (cm *connManager) Login(guestId string, userId int64, conn Connection) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.loginConns[userId] = conn
	delete(cm.guestConns, guestId)
}

func (cm *connManager) LoginOut(userId int64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.loginConns, userId)
}

func (cm *connManager) Exit(guestId string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.guestConns, guestId)
}

func (cm *connManager) GetGuest(guestId string) Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, _ := cm.guestConns[guestId]

	return conn
}

func (cm *connManager) GuestTotal() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return len(cm.guestConns)
}

func (cm *connManager) GetUser(userId int64) Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	conn, _ := cm.loginConns[userId]

	return conn
}

func (cm *connManager) UserTotal() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return len(cm.loginConns)
}

func (cm *connManager) ConnTotal() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return len(cm.guestConns) + len(cm.loginConns)
}

func (cm *connManager) AcquireConn(c net.Conn) Connection {
	conn := cm.pool.Get().(*connection)
	conn.conn = c
	return conn
}

func (cm *connManager) ReleaseConn(conn Connection) {
	conn.Reset()
	cm.pool.Put(conn)
}
