package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const (
	writeTimeout = 10 * time.Second
	pingInterval = 30 * time.Second
)

type Client struct {
	conn     *websocket.Conn
	playerID string
	lobbyCode string
	send     chan []byte
	hub      *Hub
	once     sync.Once
}

func NewClient(conn *websocket.Conn, playerID, lobbyCode string, h *Hub) *Client {
	return &Client{
		conn:      conn,
		playerID:  playerID,
		lobbyCode: lobbyCode,
		send:      make(chan []byte, 64),
		hub:       h,
	}
}

func (c *Client) PlayerID() string  { return c.playerID }
func (c *Client) LobbyCode() string { return c.lobbyCode }

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				log.Printf("[ws] клиент %s отключился штатно", c.playerID)
			} else {
				log.Printf("[ws] ошибка чтения от %s: %v", c.playerID, err)
			}
			return
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[ws] невалидный JSON от %s: %v", c.playerID, err)
			continue
		}

		c.hub.HandleMessage(c, msg)
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				log.Printf("[ws] ошибка записи клиенту %s: %v", c.playerID, err)
				return
			}
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				log.Printf("[ws] ping failed для %s: %v", c.playerID, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		log.Printf("[ws] буфер отправки переполнен для %s, отключаем", c.playerID)
		c.Close()
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.send)
	})
}
