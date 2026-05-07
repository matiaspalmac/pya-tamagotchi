package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/tamagotchi/notif/internal"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	rdb := mustRedis()
	defer rdb.Close()

	hub := internal.NewHub(rdb, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	secret := []byte(os.Getenv("JWT_SECRET"))

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		uid, err := verify(token, secret)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &internal.Client{UserID: uid, Send: make(chan []byte, 32)}
		hub.Add(client)
		go writePump(conn, client, hub)
		go readPump(conn, client, hub)
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8084"
	}
	srv := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	c, cn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cn()
	_ = srv.Shutdown(c)
}

type claims struct {
	UID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func verify(raw string, secret []byte) (uuid.UUID, error) {
	if raw == "" {
		return uuid.Nil, errors.New("no token")
	}
	t, err := jwt.ParseWithClaims(raw, &claims{}, func(t *jwt.Token) (any, error) { return secret, nil })
	if err != nil || !t.Valid {
		return uuid.Nil, errors.New("invalid")
	}
	return t.Claims.(*claims).UID, nil
}

func readPump(conn *websocket.Conn, c *internal.Client, h *internal.Hub) {
	defer func() {
		h.Remove(c)
		conn.Close()
	}()
	conn.SetReadLimit(1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func writePump(conn *websocket.Conn, c *internal.Client, h *internal.Hub) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mustRedis() *redis.Client {
	c := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})
	for i := 0; i < 30; i++ {
		if err := c.Ping(context.Background()).Err(); err == nil {
			return c
		}
		time.Sleep(time.Second)
	}
	panic("redis unreachable")
}
