package service

import (
	_interface "github.com/Point-AI/backend/internal/messenger/domain/interface"
	infrastructureInterface "github.com/Point-AI/backend/internal/messenger/service/interface"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"sync"
)

type WebSocketServiceImpl struct {
	messengerRepo infrastructureInterface.MessengerRepository
	connections   map[string]map[primitive.ObjectID]*websocket.Conn
	upgrader      *websocket.Upgrader
	mu            sync.RWMutex
}

func NewWebSocketServiceImpl(messengerRepo infrastructureInterface.MessengerRepository) _interface.WebsocketService {
	return &WebSocketServiceImpl{
		messengerRepo: messengerRepo,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make(map[string]map[primitive.ObjectID]*websocket.Conn),
	}
}

func (wss *WebSocketServiceImpl) UpgradeConnection(w http.ResponseWriter, r *http.Request, workspaceId string, userId primitive.ObjectID) (*websocket.Conn, error) {
	conn, err := wss.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	wss.AddConnection(workspaceId, conn, userId)
	return conn, nil
}

func (wss *WebSocketServiceImpl) SendToAll(workspaceId string, message []byte) {
	wss.mu.RLock()
	defer wss.mu.RUnlock()

	conns := wss.connections[workspaceId]
	if conns == nil {
		return
	}

	for id, conn := range conns {
		if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			wss.RemoveConnection(workspaceId, id)
		}
	}
}

func (wss *WebSocketServiceImpl) SendToAllButOne(workspaceId string, message []byte, userId primitive.ObjectID) {
	wss.mu.RLock()
	defer wss.mu.RUnlock()

	conns := wss.connections[workspaceId]
	if conns == nil {
		return
	}

	for id, conn := range conns {
		if id == userId {
			continue
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
			wss.RemoveConnection(workspaceId, id)
		}
	}
}

func (wss *WebSocketServiceImpl) SendToOne(message []byte, workspaceId string, userId primitive.ObjectID) {
	if err := wss.connections[workspaceId][userId].WriteMessage(websocket.BinaryMessage, message); err != nil {
		wss.RemoveConnection(workspaceId, userId)
	}
}

func (wss *WebSocketServiceImpl) AddConnection(workspaceId string, conn *websocket.Conn, userId primitive.ObjectID) {
	wss.mu.Lock()
	defer wss.mu.Unlock()

	if wss.connections[workspaceId] == nil {
		wss.connections[workspaceId] = make(map[primitive.ObjectID]*websocket.Conn)
	}

	wss.connections[workspaceId][userId] = conn
}

func (wss *WebSocketServiceImpl) RemoveConnection(workspaceId string, userId primitive.ObjectID) {
	wss.mu.Lock()
	defer wss.mu.Unlock()
	wss.connections[workspaceId][userId] = nil
}

func (wss *WebSocketServiceImpl) GetConnections(workspaceId string) map[primitive.ObjectID]*websocket.Conn {
	wss.mu.RLock()
	defer wss.mu.RUnlock()
	return wss.connections[workspaceId]
}
