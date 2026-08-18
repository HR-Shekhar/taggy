package community

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	EventMessageCreated = "message.created"
	EventMessageUpdated = "message.updated"
	EventMessageDeleted = "message.deleted"
)

func RoomKeyPod(podSlug string) string {
	return "pod:" + podSlug
}

func RoomKeyChannel(skillSlug, channelSlug string) string {
	return "channel:" + skillSlug + ":" + channelSlug
}

type RealtimeEvent struct {
	Type      string           `json:"type"`
	Message   *messageResponse `json:"message,omitempty"`
	MessageID *int64           `json:"message_id,omitempty"`
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*wsClient]struct{}
	log     zerolog.Logger
	upgrader websocket.Upgrader
}

type wsClient struct {
	hub  *Hub
	room string
	conn *websocket.Conn
	send chan []byte
}

func NewHub(log zerolog.Logger, allowedOrigins []string) *Hub {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return &Hub{
		rooms: make(map[string]map[*wsClient]struct{}),
		log:   log,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				if len(originSet) == 0 {
					return true
				}
				_, ok := originSet[origin]
				return ok
			},
		},
	}
}

func (h *Hub) Publish(room string, event RealtimeEvent) {
	payload, err := json.Marshal(event)
	if err != nil {
		h.log.Error().Err(err).Str("room", room).Msg("chat hub marshal failed")
		return
	}

	h.mu.RLock()
	clients := h.rooms[room]
	for client := range clients {
		select {
		case client.send <- payload:
		default:
			h.log.Warn().Str("room", room).Msg("chat hub client send buffer full")
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) subscribe(room string, client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*wsClient]struct{})
	}
	h.rooms[room][client] = struct{}{}
}

func (h *Hub) unsubscribe(client *wsClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[client.room]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.rooms, client.room)
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(25 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *wsClient) readPump() {
	defer func() {
		c.hub.unsubscribe(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
