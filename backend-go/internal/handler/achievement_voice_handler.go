package handler

import (
	"bytes"
	"log"
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// ═══ فويس مسج للإنجاز — طلب (ع) 10-05 ═══
// الموظف يسجّل صوته ويا الإنجاز، والمدير يسمعه. الصوت يبقى بتخزيننا —
// ما يروح لأي مزوّد (قاعدة «ماكو بيانات شخصية لمزوّد خارجي») لحد ما (ع)
// يقرر تحليله. مسار خاص مو /api/files: الصوت بيانات شخصية، فيفتحه صاحبه
// والمدير بس.

const maxVoiceBytes = 3 << 20 // ٣ ميغا ≈ ٣ دقايق opus
const maxVoiceSeconds = 180

func (h *AchievementHandler) SetVoice(v *repository.AchievementVoiceRepository, s storage.Store, analyzer *service.VoiceAnalysisService) {
	h.voices, h.store, h.analyzer = v, s, analyzer
}

func (h *AchievementHandler) attachVoices(rows []model.Achievement) {
	if h.voices == nil || len(rows) == 0 {
		return
	}
	ids := make([]string, len(rows))
	for i := range rows {
		ids[i] = rows[i].ID
	}
	secs := h.voices.Seconds(ids)
	for i := range rows {
		if v, ok := secs[rows[i].ID]; ok {
			s := v
			rows[i].VoiceSeconds = &s
		}
	}
}

// sniffAudio نوع الصوت من بصمته: webm/ogg (أندرويد وكروم)، mp4 (آيفون)، wav.
func sniffAudio(b []byte) string {
	switch {
	case len(b) >= 4 && bytes.Equal(b[:4], []byte{0x1A, 0x45, 0xDF, 0xA3}):
		return "audio/webm"
	case len(b) >= 4 && bytes.Equal(b[:4], []byte("OggS")):
		return "audio/ogg"
	case len(b) >= 12 && bytes.Equal(b[4:8], []byte("ftyp")):
		return "audio/mp4"
	case len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WAVE")):
		return "audio/wav"
	}
	return ""
}

func voiceExt(ct string) string {
	switch ct {
	case "audio/webm":
		return ".webm"
	case "audio/ogg":
		return ".ogg"
	case "audio/mp4":
		return ".m4a"
	}
	return ".wav"
}

// POST /api/achievements/{id}/voice — multipart: file + seconds. صاحب الإنجاز بس.
func (h *AchievementHandler) UploadVoice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	me := middleware.EmployeeIDFromContext(r)
	a, err := h.service.Repo().FindByID(id)
	if err != nil || a == nil {
		WriteError(w, http.StatusNotFound, "الإنجاز مو موجود")
		return
	}
	if a.EmployeeID != me {
		WriteError(w, http.StatusForbidden, "تگدر ترفق فويس لإنجازك إنت بس")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxVoiceBytes+64<<10)
	if err := r.ParseMultipartForm(maxVoiceBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "الفويس أطول من المسموح (٣ دقايق)")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو فويس بالطلب")
		return
	}
	defer file.Close()
	data, err := storage.ReadLimited(file, maxVoiceBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	ct := sniffAudio(data)
	if ct == "" {
		WriteError(w, http.StatusBadRequest, "صيغة الصوت مو مدعومة")
		return
	}
	secs, _ := strconv.Atoi(r.FormValue("seconds"))
	if secs < 0 || secs > maxVoiceSeconds {
		secs = maxVoiceSeconds
	}
	key := "voice/" + id + voiceExt(ct)
	if err := h.store.Put(r.Context(), key, data, ct); err != nil {
		log.Printf("voice upload: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الفويس")
		return
	}
	if err := h.voices.Save(repository.AchievementVoice{AchievementID: id, FileKey: key, ContentType: ct, Seconds: secs}); err != nil {
		log.Printf("voice save: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الفويس")
		return
	}
	if h.analyzer != nil {
		h.analyzer.Kick(id)
	}
	WriteJSON(w, http.StatusCreated, map[string]any{"ok": true, "seconds": secs, "analyzing": h.analyzer != nil && h.analyzer.Enabled()})
}

// GET /api/achievements/{id}/voice — صاحبه أو المدير/المالك.
func (h *AchievementHandler) Voice(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a, err := h.service.Repo().FindByID(id)
	if err != nil || a == nil {
		WriteError(w, http.StatusNotFound, "الإنجاز مو موجود")
		return
	}
	if a.EmployeeID != middleware.EmployeeIDFromContext(r) && middleware.RoleFromContext(r) != "ADMIN" && middleware.RoleFromContext(r) != "OWNER" {
		WriteError(w, http.StatusForbidden, "الفويس لصاحبه والمدير بس")
		return
	}
	v, err := h.voices.Get(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "ماكو فويس")
		return
	}
	data, ct, err := h.store.Get(r.Context(), v.FileKey)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الفويس مو موجود بالتخزين")
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, no-store")
	_, _ = w.Write(data)
}

// GET /api/achievements/{id}/voice/analysis — شنو فهم ماتركس من الفويس. صاحبه أو المدير.
func (h *AchievementHandler) VoiceAnalysis(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a, err := h.service.Repo().FindByID(id)
	if err != nil || a == nil {
		WriteError(w, http.StatusNotFound, "الإنجاز مو موجود")
		return
	}
	role := middleware.RoleFromContext(r)
	if a.EmployeeID != middleware.EmployeeIDFromContext(r) && role != "ADMIN" && role != "OWNER" {
		WriteError(w, http.StatusForbidden, "التحليل لصاحبه والمدير بس")
		return
	}
	v, err := h.voices.Analysis(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "ماكو فويس")
		return
	}
	enabled := h.analyzer != nil && h.analyzer.Enabled()
	WriteJSON(w, http.StatusOK, map[string]any{"analysis": v, "enabled": enabled})
}

// GET /api/ai/dialect — قاموس اللهجة الي تعلّمه ماتركس. المدير والمالك.
func (h *AchievementHandler) Dialect(w http.ResponseWriter, r *http.Request) {
	rows, err := h.voices.Phrases(200)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب القاموس")
		return
	}
	done, pending := h.voices.AnalyzedCount()
	enabled := h.analyzer != nil && h.analyzer.Enabled()
	WriteJSON(w, http.StatusOK, map[string]any{"phrases": rows, "analyzed": done, "pending": pending, "enabled": enabled})
}
