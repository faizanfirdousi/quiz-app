package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"

	"kahootclone/internal/db"
	"kahootclone/internal/models"
	"kahootclone/internal/observability"
)

// Broadcaster handles sending WebSocket messages to connections.
// In local mode, it uses the gorilla/websocket Hub.
// In production mode, it would use the API Gateway Management API.
type Broadcaster struct {
	DB  *db.Client
	Hub *Hub // non-nil in local mode
	Env string
}

// NewBroadcaster creates a new Broadcaster.
func NewBroadcaster(dbClient *db.Client, env string) *Broadcaster {
	return &Broadcaster{
		DB:  dbClient,
		Env: env,
	}
}

// SetHub sets the local WebSocket hub for local development.
func (b *Broadcaster) SetHub(hub *Hub) {
	b.Hub = hub
}

// SendToConnection sends a WS message to a single connectionId.
func (b *Broadcaster) SendToConnection(ctx context.Context, connectionID string, payload models.WSOutbound) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if b.Env == "local" && b.Hub != nil {
		return b.Hub.SendToConnection(connectionID, data)
	}

	// Production: use API Gateway Management API
	// This would be implemented with apigatewaymanagementapi.PostToConnection
	observability.Warn(ctx, "production broadcast not implemented", "connectionId", connectionID)
	return nil
}

// BroadcastToSession sends a message to all connections in a session.
func (b *Broadcaster) BroadcastToSession(ctx context.Context, sessionID string, payload models.WSOutbound) error {
	observability.Debug(ctx, "broadcasting to session", "sessionId", sessionID, "type", payload.Type)

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if b.Env == "local" && b.Hub != nil {
		return b.Hub.BroadcastToSession(sessionID, data)
	}

	// Production: fetch connections from DynamoDB and post to each
	connections, err := b.DB.GetConnectionsBySession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get connections: %w", err)
	}

	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			if sendErr := b.SendToConnection(ctx, cid, payload); sendErr != nil {
				observability.Warn(ctx, "failed to send to connection", "connectionId", cid, "error", sendErr.Error())
				// Stale connection (410 Gone) — delete from DynamoDB
				_ = b.DB.DeleteConnection(ctx, sessionID, cid)
			}
		}(conn.ConnectionID)
	}
	wg.Wait()

	return nil
}

// SendToPlayer sends a message to a specific player in a session.
func (b *Broadcaster) SendToPlayer(ctx context.Context, sessionID, userID string, payload models.WSOutbound) error {
	conn, err := b.DB.GetConnectionByUserID(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to find player connection: %w", err)
	}
	return b.SendToConnection(ctx, conn.ConnectionID, payload)
}

// --- Local WebSocket Hub ---

// Hub manages local WebSocket connections using gorilla/websocket.
type Hub struct {
	mu          sync.RWMutex
	connections map[string]*Connection         // connectionId -> Connection
	sessions    map[string]map[string]struct{} // sessionId -> set of connectionIds
}

// Connection wraps a gorilla/websocket connection.
type Connection struct {
	ID        string
	SessionID string
	Conn      *websocket.Conn
	mu        sync.Mutex
}

// NewHub creates a new WebSocket Hub.
func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]*Connection),
		sessions:    make(map[string]map[string]struct{}),
	}
}

// Register adds a connection to the hub.
func (h *Hub) Register(connectionID, sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connections[connectionID] = &Connection{
		ID:        connectionID,
		SessionID: sessionID,
		Conn:      conn,
	}

	if h.sessions[sessionID] == nil {
		h.sessions[sessionID] = make(map[string]struct{})
	}
	h.sessions[sessionID][connectionID] = struct{}{}

	slog.Info("WS connection registered", "connectionId", connectionID, "sessionId", sessionID)
}

// Unregister removes a connection from the hub.
func (h *Hub) Unregister(connectionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conn, ok := h.connections[connectionID]
	if !ok {
		return
	}

	delete(h.connections, connectionID)
	if sessionConns, ok := h.sessions[conn.SessionID]; ok {
		delete(sessionConns, connectionID)
		if len(sessionConns) == 0 {
			delete(h.sessions, conn.SessionID)
		}
	}

	slog.Info("WS connection unregistered", "connectionId", connectionID, "sessionId", conn.SessionID)
}

// SendToConnection sends a message to a specific connection.
func (h *Hub) SendToConnection(connectionID string, data []byte) error {
	h.mu.RLock()
	conn, ok := h.connections[connectionID]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("connection %s not found in hub", connectionID)
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	return conn.Conn.WriteMessage(websocket.TextMessage, data)
}

// BroadcastToSession sends a message to all connections in a session.
func (h *Hub) BroadcastToSession(sessionID string, data []byte) error {
	h.mu.RLock()
	connIDs, ok := h.sessions[sessionID]
	if !ok {
		h.mu.RUnlock()
		return nil
	}
	// Copy IDs to avoid holding lock during sends
	ids := make([]string, 0, len(connIDs))
	for id := range connIDs {
		ids = append(ids, id)
	}
	h.mu.RUnlock()

	var lastErr error
	for _, id := range ids {
		if err := h.SendToConnection(id, data); err != nil {
			lastErr = err
			slog.Warn("failed to send to connection in broadcast", "connectionId", id, "error", err.Error())
		}
	}
	return lastErr
}
