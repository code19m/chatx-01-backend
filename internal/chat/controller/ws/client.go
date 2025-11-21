package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

	// Size of the send channel buffer.
	sendBufferSize = 256
)

// Client represents a single WebSocket connection.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	userID  int
	chatIDs []int
	send    chan *Event
	logger  *slog.Logger

	closeOnce sync.Once
	closed    chan struct{}
}

// NewClient creates a new Client instance.
func NewClient(hub *Hub, conn *websocket.Conn, userID int, chatIDs []int, logger *slog.Logger) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		userID:  userID,
		chatIDs: chatIDs,
		send:    make(chan *Event, sendBufferSize),
		logger:  logger,
		closed:  make(chan struct{}),
	}
}

// UserID returns the user ID associated with this client.
func (c *Client) UserID() int {
	return c.userID
}

// ChatIDs returns the chat IDs the user is participating in.
func (c *Client) ChatIDs() []int {
	return c.chatIDs
}

// Run starts the client's read and write pumps.
// This blocks until the connection is closed.
func (c *Client) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		c.writePump(ctx)
	}()

	go func() {
		defer wg.Done()
		c.readPump(ctx)
	}()

	wg.Wait()
}

// Close closes the client connection.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closed)
		c.conn.Close(websocket.StatusNormalClosure, "connection closed")
	})
}

// Send queues an event to be sent to the client.
func (c *Client) Send(event *Event) bool {
	select {
	case c.send <- event:
		return true
	case <-c.closed:
		return false
	default:
		// Buffer full
		return false
	}
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
	}()

	c.conn.SetReadLimit(maxMessageSize)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closed:
			return
		default:
		}

		// Read with timeout
		readCtx, cancel := context.WithTimeout(ctx, pongWait)
		var msg ClientMessage
		err := wsjson.Read(readCtx, c.conn, &msg)
		cancel()

		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				c.logger.Debug("client disconnected normally", "user_id", c.userID)
			} else {
				c.logger.Debug("read error", "user_id", c.userID, "error", err)
			}
			return
		}

		c.handleMessage(&msg)
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.hub.Unregister(c)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closed:
			return

		case event, ok := <-c.send:
			if !ok {
				// Channel closed
				return
			}

			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := wsjson.Write(writeCtx, c.conn, event)
			cancel()

			if err != nil {
				c.logger.Debug("write error", "user_id", c.userID, "error", err)
				return
			}

		case <-ticker.C:
			// Send ping
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()

			if err != nil {
				c.logger.Debug("ping error", "user_id", c.userID, "error", err)
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client.
func (c *Client) handleMessage(msg *ClientMessage) {
	switch msg.Type {
	case EventTypingStart, EventTypingStop:
		c.handleTyping(msg)
	default:
		c.logger.Debug("unknown message type", "type", msg.Type, "user_id", c.userID)
	}
}

// handleTyping broadcasts typing events to chat participants.
func (c *Client) handleTyping(msg *ClientMessage) {
	if msg.Payload.ChatID == 0 {
		return
	}

	// Verify user is participant in the chat
	isParticipant := false
	for _, chatID := range c.chatIDs {
		if chatID == msg.Payload.ChatID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.logger.Warn("user not participant in chat",
			"user_id", c.userID,
			"chat_id", msg.Payload.ChatID,
		)
		return
	}

	event := &Event{
		Type: msg.Type,
		Payload: TypingPayload{
			ChatID: msg.Payload.ChatID,
			UserID: c.userID,
		},
	}

	c.hub.BroadcastToChat(msg.Payload.ChatID, event, c.userID)
}

// MarshalJSON implements json.Marshaler for Event.
func (e *Event) MarshalJSON() ([]byte, error) {
	type eventAlias Event
	return json.Marshal((*eventAlias)(e))
}
