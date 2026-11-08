package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// 🔔 اشتراك الأجهزة بإشعارات الموبايل (ع) 11-08 — كل موظف لنفسه بس.
type PushHandler struct{ push *service.WebPush }

func NewPushHandler(p *service.WebPush) *PushHandler { return &PushHandler{push: p} }

// GET /api/push/key
func (h *PushHandler) Key(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"publicKey": h.push.PublicKey()})
}

type pushSubIn struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// POST /api/push/subscribe
func (h *PushHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var in pushSubIn
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.push.Subscribe(middleware.EmployeeIDFromContext(r), in.Endpoint, in.Keys.P256dh, in.Keys.Auth, r.UserAgent()); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/push/unsubscribe
func (h *PushHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var in pushSubIn
	_ = DecodeJSON(r, &in)
	_ = h.push.Unsubscribe(middleware.EmployeeIDFromContext(r), in.Endpoint)
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
