package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"staffmange-api/internal/service"
)

// SecretarySpeakHandler 🔊 كمال يحچي (دازاي 11-17): POST /api/secretary/speak {text} ← mp3.
// بلا مفتاح أو صوت = 404 (الميزة مو مربوطة).
type SecretarySpeakHandler struct{ tts *service.ElevenLabsTTS }

func NewSecretarySpeakHandler(tts *service.ElevenLabsTTS) *SecretarySpeakHandler {
	return &SecretarySpeakHandler{tts: tts}
}

func (h *SecretarySpeakHandler) Speak(w http.ResponseWriter, r *http.Request) {
	if h.tts == nil {
		WriteError(w, http.StatusNotFound, service.ErrTTSOff.Error())
		return
	}
	var in struct {
		Text string `json:"text"`
	}
	if err := DecodeJSON(r, &in); err != nil || service.SpeakText(in.Text) == "" {
		WriteError(w, http.StatusBadRequest, "ماكو كلام ينگال")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()
	b, err := h.tts.Speak(ctx, in.Text)
	if err != nil {
		if errors.Is(err, service.ErrTTSOff) {
			WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("[secretary speak] %v", err)
		WriteError(w, http.StatusBadGateway, "تعذّر تحويل الرد لصوت")
		return
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}
