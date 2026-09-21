package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"server-monitoring/internal/models"

	"github.com/gorilla/websocket"
)

const (
	wsWriteWait      = 10 * time.Second
	wsPongWait       = 60 * time.Second
	wsPingPeriod     = (wsPongWait * 9) / 10
	wsMaxMessageSize = 1024
	wsClientBuffer   = 64
	wsBroadcastBuf   = 64
)

// NotificationEnvelope is the wire shape pushed over the socket. The
// `type` discriminator lets clients multiplex future message kinds
// (heartbeats, acks, …) without inspecting the payload.
type NotificationEnvelope struct {
	Type         string               `json:"type"`
	Notification *models.Notification `json:"notification,omitempty"`
}

// NotificationHub fans notifications out to every connected socket. A
// single goroutine owns the client set, so registration, removal and
// broadcast never race.
type NotificationHub interface {
	Register(client *NotificationClient)
	Unregister(client *NotificationClient)
	Broadcast(n models.Notification)
	Close()
}

type notificationHub struct {
	clients    map[*NotificationClient]struct{}
	register   chan *NotificationClient
	unregister chan *NotificationClient
	broadcast  chan []byte
	done       chan struct{}
	closeOnce  sync.Once
}

func NewNotificationHub() NotificationHub {
	h := &notificationHub{
		clients:    make(map[*NotificationClient]struct{}),
		register:   make(chan *NotificationClient),
		unregister: make(chan *NotificationClient),
		broadcast:  make(chan []byte, wsBroadcastBuf),
		done:       make(chan struct{}),
	}
	go h.run()
	return h
}

func (h *notificationHub) run() {
	for {
		select {
		case <-h.done:
			for c := range h.clients {
				close(c.send)
				delete(h.clients, c)
			}
			return
		case c := <-h.register:
			h.clients[c] = struct{}{}
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// Slow client: drop it instead of blocking the hub.
					close(c.send)
					delete(h.clients, c)
				}
			}
		}
	}
}

func (h *notificationHub) Register(client *NotificationClient) {
	select {
	case h.register <- client:
	case <-h.done:
	}
}

func (h *notificationHub) Unregister(client *NotificationClient) {
	select {
	case h.unregister <- client:
	case <-h.done:
	}
}

func (h *notificationHub) Broadcast(n models.Notification) {
	msg, err := json.Marshal(NotificationEnvelope{Type: "notification", Notification: &n})
	if err != nil {
		log.Printf("notifications: marshal failed: %v", err)
		return
	}
	select {
	case h.broadcast <- msg:
	case <-h.done:
	}
}

func (h *notificationHub) Close() {
	h.closeOnce.Do(func() { close(h.done) })
}

// NotificationClient is one browser's WebSocket connection. All writes
// go through `send` so a slow reader can't block the hub; the write
// pump drains it and keeps the link alive with periodic pings.
type NotificationClient struct {
	hub  NotificationHub
	conn *websocket.Conn
	send chan []byte
}

func NewNotificationClient(conn *websocket.Conn, hub NotificationHub) *NotificationClient {
	return &NotificationClient{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, wsClientBuffer),
	}
}

// Run starts the read/write pumps and returns immediately. They run
// until the connection closes.
func (c *NotificationClient) Run() {
	go c.writePump()
	go c.readPump()
}

// readPump exists to notice disconnects and to keep the read deadline
// fresh via pongs. Inbound messages are ignored.
func (c *NotificationClient) readPump() {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(wsMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *NotificationClient) writePump() {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
