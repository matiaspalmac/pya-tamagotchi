package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

	"github.com/tamagotchi/pet/internal/handler"
	"github.com/tamagotchi/pet/internal/repo"
	"github.com/tamagotchi/pet/internal/service"
)

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func main() {
	db := mustDB()
	defer db.Close()
	rdb := mustRedis()
	defer rdb.Close()

	petRepo := repo.NewPetRepo(db)
	svc := service.New(petRepo, rdb, envInt("TICK_INTERVAL_SEC", 30))

	secret := []byte(os.Getenv("JWT_SECRET"))
	authFn := func(r *http.Request) (uuid.UUID, error) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if raw == "" {
			return uuid.Nil, errors.New("no token")
		}
		type claims struct {
			UID uuid.UUID `json:"uid"`
			jwt.RegisteredClaims
		}
		t, err := jwt.ParseWithClaims(raw, &claims{}, func(t *jwt.Token) (any, error) {
			return secret, nil
		})
		if err != nil || !t.Valid {
			return uuid.Nil, errors.New("invalid token")
		}
		c := t.Claims.(*claims)
		return c.UID, nil
	}

	h := handler.New(svc, authFn)

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Timeout(15*time.Second))
	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"*"}}))
	h.Routes(r)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8082"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
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
