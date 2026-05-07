package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/tamagotchi/social/internal"
)

func main() {
	db := mustDB()
	defer db.Close()
	rdb := mustRedis()
	defer rdb.Close()

	svc := internal.New(db, rdb)
	secret := []byte(os.Getenv("JWT_SECRET"))

	auth := func(r *http.Request) (uuid.UUID, error) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if raw == "" {
			return uuid.Nil, errors.New("no token")
		}
		type cl struct {
			UID uuid.UUID `json:"uid"`
			jwt.RegisteredClaims
		}
		t, err := jwt.ParseWithClaims(raw, &cl{}, func(t *jwt.Token) (any, error) { return secret, nil })
		if err != nil || !t.Valid {
			return uuid.Nil, errors.New("invalid token")
		}
		return t.Claims.(*cl).UID, nil
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Timeout(15*time.Second))
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"*"}}))

	r.Post("/friends/request", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		var req struct{ ToUser uuid.UUID `json:"to_user"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpErr(w, 400, "bad body")
			return
		}
		f, err := svc.Request(r.Context(), uid, req.ToUser)
		if err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 201, f)
	})

	r.Post("/friends/accept", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		var req struct{ RequestID uuid.UUID `json:"request_id"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpErr(w, 400, "bad body")
			return
		}
		f, err := svc.Accept(r.Context(), uid, req.RequestID)
		if err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, f)
	})

	r.Get("/friends", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		fs, err := svc.ListFriends(r.Context(), uid)
		if err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, fs)
	})

	r.Post("/gifts", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		var req struct {
			ToUser uuid.UUID `json:"to_user"`
			Type   string    `json:"type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpErr(w, 400, "bad body")
			return
		}
		g, err := svc.SendGift(r.Context(), uid, req.ToUser, req.Type)
		if err != nil {
			httpErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 201, g)
	})

	r.Get("/gifts/inbox", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		gs, err := svc.Inbox(r.Context(), uid)
		if err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, gs)
	})

	r.Post("/gifts/{id}/claim", func(w http.ResponseWriter, r *http.Request) {
		uid, err := auth(r)
		if err != nil {
			httpErr(w, 401, err.Error())
			return
		}
		gid, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			httpErr(w, 400, "bad id")
			return
		}
		g, err := svc.Claim(r.Context(), uid, gid)
		if err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, g)
	})

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8083"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r, ReadHeaderTimeout: 5 * time.Second}

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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func mustDB() *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	var db *sqlx.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sqlx.Open("postgres", dsn)
		if err == nil {
			if err = db.Ping(); err == nil {
				return db
			}
		}
		time.Sleep(time.Second)
	}
	panic(err)
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
