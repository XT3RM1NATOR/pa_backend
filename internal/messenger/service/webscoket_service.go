package service

import (
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	"github.com/gorilla/websocket"
	"sync"
)

type WebSocketServiceImpl struct {
	connections map[string]*websocket.Conn
	upgrader    websocket.Upgrader
	mu          sync.RWMutex
}

func NewWebSocketServiceImpl() _interface.WebsocketService {
	return &WebSocketServiceImpl{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[string]*websocket.Conn),
	}
}

func (wss *WebSocketServiceImpl) AddConnection(id string, conn *websocket.Conn) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	wss.connections[id] = conn
}

func (wss *WebSocketServiceImpl) RemoveConnection(id string) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	delete(wss.connections, id)
}

func (wss *WebSocketServiceImpl) GetConnection(id string) (*websocket.Conn, bool) {
	wss.mu.RLock()
	defer wss.mu.RUnlock()
	conn, ok := wss.connections[id]
	return conn, ok
}
