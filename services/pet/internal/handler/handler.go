package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/tamagotchi/pet/internal/repo"
	"github.com/tamagotchi/pet/internal/service"
)

type Authn func(*http.Request) (uuid.UUID, error)

type H struct {
	svc  *service.Service
	auth Authn
}

func New(s *service.Service, a Authn) *H { return &H{svc: s, auth: a} }

func (h *H) Routes(r chi.Router) {
	r.Post("/pets", h.create)
	r.Get("/pets/mine", h.mine)
	r.Get("/pets/{id}", h.get)
	r.Post("/pets/{id}/feed", h.action(h.svc.Feed))
	r.Post("/pets/{id}/play", h.action(h.svc.Play))
	r.Post("/pets/{id}/sleep", h.action(h.svc.Sleep))
	r.Post("/pets/{id}/heal", h.action(h.svc.Heal))
	r.Get("/pets/{id}/events", h.events)
}

type actionFn func(ctx context.Context, owner, id uuid.UUID) (*repo.Pet, error)

func (h *H) action(fn actionFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := h.auth(r)
		if err != nil {
			writeErr(w, 401, "unauthorized")
			return
		}
		pid, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			writeErr(w, 400, "bad id")
			return
		}
		p, err := fn(r.Context(), uid, pid)
		if err != nil {
			writeErr(w, statusFor(err), err.Error())
			return
		}
		writeJSON(w, 200, p)
	}
}

type createReq struct {
	Name    string `json:"name"`
	Species string `json:"species"`
}

func (h *H) create(w http.ResponseWriter, r *http.Request) {
	uid, err := h.auth(r)
	if err != nil {
		writeErr(w, 401, "unauthorized")
		return
	}
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeErr(w, 400, "invalid body")
		return
	}
	if req.Species == "" {
		req.Species = "blob"
	}
	p, err := h.svc.Create(r.Context(), uid, req.Name, req.Species)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, p)
}

func (h *H) mine(w http.ResponseWriter, r *http.Request) {
	uid, err := h.auth(r)
	if err != nil {
		writeErr(w, 401, "unauthorized")
		return
	}
	ps, err := h.svc.Mine(r.Context(), uid)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, ps)
}

func (h *H) get(w http.ResponseWriter, r *http.Request) {
	uid, err := h.auth(r)
	if err != nil {
		writeErr(w, 401, "unauthorized")
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	p, err := h.svc.Get(r.Context(), uid, pid)
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, 200, p)
}

func (h *H) events(w http.ResponseWriter, r *http.Request) {
	uid, err := h.auth(r)
	if err != nil {
		writeErr(w, 401, "unauthorized")
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, 400, "bad id")
		return
	}
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	evs, err := h.svc.Events(r.Context(), uid, pid, limit)
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, 200, evs)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, service.ErrCooldown):
		return 429
	case errors.Is(err, service.ErrLowEnergy):
		return 409
	case errors.Is(err, service.ErrDead):
		return 410
	case errors.Is(err, service.ErrNotOwner):
		return 403
	case errors.Is(err, repo.ErrNotFound):
		return 404
	default:
		return 500
	}
}
