package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	UserID uuid.UUID
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
	rdb     *redis.Client
	log     *slog.Logger
}

func NewHub(rdb *redis.Client, log *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]struct{}),
		rdb:     rdb,
		log:     log,
	}
}

func (h *Hub) Add(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.UserID]; !ok {
		h.clients[c.UserID] = make(map[*Client]struct{})
	}
	h.clients[c.UserID][c] = struct{}{}
}

func (h *Hub) Remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.UserID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, c.UserID)
		}
	}
	close(c.Send)
}

func (h *Hub) BroadcastUser(uid uuid.UUID, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[uid] {
		select {
		case c.Send <- msg:
		default:
		}
	}
}

func (h *Hub) BroadcastAll(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.clients {
		for c := range set {
			select {
			case c.Send <- msg:
			default:
			}
		}
	}
}

// Run subscribes to Redis channels and fans out to clients.
func (h *Hub) Run(ctx context.Context) {
	pubsub := h.rdb.Subscribe(ctx, "pet:tick", "pet:event", "pet:evolved", "pet:created", "gift:received", "friend:request")
	defer pubsub.Close()
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.dispatch(msg.Channel, []byte(msg.Payload))
		}
	}
}

func (h *Hub) dispatch(channel string, payload []byte) {
	var generic map[string]any
	if err := json.Unmarshal(payload, &generic); err != nil {
		h.log.Warn("bad payload", "ch", channel, "err", err)
		return
	}
	envelope, _ := json.Marshal(map[string]any{
		"type": channel,
		"data": generic,
	})
	// MVP: broadcast to all. Mejora futuro: routing por owner_id en payload.
	if uidStr, ok := generic["owner_id"].(string); ok {
		if uid, err := uuid.Parse(uidStr); err == nil {
			h.BroadcastUser(uid, envelope)
			return
		}
	}
	h.BroadcastAll(envelope)
}
