package service

import (
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type WebSocketServiceImpl struct {
	messengerRepo infrastructureInterface.MessengerRepository
	connections   map[string][]*websocket.Conn
	upgrader      websocket.Upgrader
	mu            sync.RWMutex
}

func NewWebSocketServiceImpl(messengerRepo infrastructureInterface.MessengerRepository) _interface.WebsocketService {
	return &WebSocketServiceImpl{
		messengerRepo: messengerRepo,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[string][]*websocket.Conn),
	}
}

func (wss *WebSocketServiceImpl) UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string) (*websocket.Conn, error) {
	conn, err := wss.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	wss.AddConnection(workspaceId, conn)
	return conn, err
}

func (wss *WebSocketServiceImpl) SendToAll(workspaceId string, message []byte) {
	wss.mu.RLock()
	defer wss.mu.RUnlock()

	conns := wss.connections[workspaceId]
	if conns == nil {
		return
	}

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// Handle error (e.g., connection closed)
			// You may choose to remove the connection or log the error
		}
	}
}

func (wss *WebSocketServiceImpl) AddConnection(workspaceId string, conn *websocket.Conn) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	wss.connections[workspaceId] = append(wss.connections[workspaceId], conn)
}

func (wss *WebSocketServiceImpl) RemoveConnection(workspaceId string, conn *websocket.Conn) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	conns, ok := wss.connections[workspaceId]
	if !ok {
		return
	}
	for i, c := range conns {
		if c == conn {
			c.Close()

			wss.connections[workspaceId] = append(conns[:i], conns[i+1:]...)
			return
		}
	}
}

func (wss *WebSocketServiceImpl) GetConnections(workspaceId string) []*websocket.Conn {
	wss.mu.RLock()
	defer wss.mu.RUnlock()
	return wss.connections[workspaceId]
}
