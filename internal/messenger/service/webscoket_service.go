package service

import (
	"golang.org/x/net/websocket"
	"sync"
)

type WebSocketService struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		connections: make(map[string]*websocket.Conn),
	}
}

func (m *WebSocketService) AddConnection(id string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[id] = conn
}

func (m *WebSocketService) RemoveConnection(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, id)
}

func (m *WebSocketService) GetConnection(id string) (*websocket.Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[id]
	return conn, ok
}
