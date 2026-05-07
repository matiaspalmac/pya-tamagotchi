package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type claims struct {
	UID uuid.UUID `json:"uid"`
	jwt.RegisteredClaims
}

func main() {
	authURL := mustURL("AUTH_URL")
	petURL := mustURL("PET_URL")
	socialURL := mustURL("SOCIAL_URL")
	notifURL := mustURL("NOTIF_URL")

	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	petProxy := httputil.NewSingleHostReverseProxy(petURL)
	socialProxy := httputil.NewSingleHostReverseProxy(socialURL)
	notifProxy := httputil.NewSingleHostReverseProxy(notifURL)

	secret := []byte(os.Getenv("JWT_SECRET"))

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Recoverer, middleware.Logger)
	allowed := []string{"http://localhost:5173"}
	if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
		allowed = append(allowed, strings.Split(extra, ",")...)
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowed,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
	}))
	r.Use(httprate.LimitByIP(100, time.Minute))

	// public
	r.Mount("/auth/register", authProxy)
	r.Mount("/auth/login", authProxy)
	r.Mount("/auth/refresh", authProxy)

	// protected
	protected := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			raw := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
			if raw == "" {
				httpErr(w, 401, "missing token")
				return
			}
			t, err := jwt.ParseWithClaims(raw, &claims{}, func(t *jwt.Token) (any, error) { return secret, nil })
			if err != nil || !t.Valid {
				httpErr(w, 401, "invalid token")
				return
			}
			next.ServeHTTP(w, req)
		})
	}

	r.Group(func(r chi.Router) {
		r.Use(protected)
		r.Mount("/auth/me", authProxy)
		r.Mount("/pets", petProxy)
		r.Mount("/friends", socialProxy)
		r.Mount("/gifts", socialProxy)
	})

	// WS no usa middleware HTTP estándar; auth en query token (notif valida)
	r.Get("/ws", notifProxy.ServeHTTP)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
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

func mustURL(env string) *url.URL {
	v := os.Getenv(env)
	if v == "" {
		panic(env + " not set")
	}
	u, err := url.Parse(v)
	if err != nil {
		panic(err)
	}
	return u
}

func httpErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
