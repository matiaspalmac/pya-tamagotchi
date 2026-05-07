package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/tamagotchi/auth/internal/service"
)

type H struct{ svc *service.Service }

func New(s *service.Service) *H { return &H{svc: s} }

type registerReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *H) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	if req.Email == "" || req.Username == "" || len(req.Password) < 8 {
		writeErr(w, 400, "missing fields or weak password")
		return
	}
	u, tp, err := h.svc.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		writeErr(w, 409, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"user": u, "tokens": tp})
}

func (h *H) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	u, tp, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeErr(w, 401, "invalid credentials")
		return
	}
	writeJSON(w, 200, map[string]any{"user": u, "tokens": tp})
}

func (h *H) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	tp, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		writeErr(w, 401, "invalid refresh")
		return
	}
	writeJSON(w, 200, map[string]any{"tokens": tp})
}

func (h *H) Me(w http.ResponseWriter, r *http.Request) {
	tok := bearer(r)
	if tok == "" {
		writeErr(w, 401, "missing token")
		return
	}
	c, err := h.svc.Verify(tok)
	if err != nil {
		writeErr(w, 401, "invalid token")
		return
	}
	u, err := h.svc.GetUser(r.Context(), c.UserID)
	if err != nil {
		writeErr(w, 404, "user not found")
		return
	}
	writeJSON(w, 200, map[string]any{"user": u})
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}

var _ = errors.New
