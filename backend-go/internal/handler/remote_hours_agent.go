package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ═══ 🤖 ساعات Claude لسجل دوام دازاي (دازاي 11-06، قرار 0011) ═══
// Claude يسجّل وقت شغله (الفحص + القراءة + التصليح) مباشرة بدال ما دازاي ينقله بإيده.
// الرمز (REMOTE_HOURS_AGENT_TOKEN) يسجّل **لموظف واحد ثابت** من الإعدادات، والموظف ما ينرسل
// بالطلب. حدود: ساعتين بالطلب، ٦ ساعات للوكيل باليوم، يومين لورا بالكثير، ومفتاح ما يتكرر.

const (
	agentMaxPerCall = 2 * 3600
	agentMaxPerDay  = 6 * 3600
	agentMaxDaysAgo = 2
)

var agentKeyRe = regexp.MustCompile(`^[A-Za-z0-9._:-]{6,64}$`)

// SetAgent يفعّل مسار الوكيل. رمز أقصر من ٣٢ حرف أو بلا موظف = مطفي (٤٠٤).
func (h *RemoteHoursHandler) SetAgent(token, employeeID string) {
	if len(token) >= 32 && strings.TrimSpace(employeeID) != "" {
		h.agentToken, h.agentEmployee = token, strings.TrimSpace(employeeID)
	}
}

type agentHoursBody struct {
	Date    string `json:"date"`
	Minutes int    `json:"minutes"`
	Seconds int    `json:"seconds"`
	Note    string `json:"note"`
}

// POST /api/remote-hours/agent — ترويسة X-Remote-Hours-Agent-Token وIdempotency-Key.
func (h *RemoteHoursHandler) AgentAdd(w http.ResponseWriter, r *http.Request) {
	got := r.Header.Get("X-Remote-Hours-Agent-Token")
	if h.agentToken == "" || subtle.ConstantTimeCompare([]byte(got), []byte(h.agentToken)) != 1 {
		http.NotFound(w, r)
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if !agentKeyRe.MatchString(key) {
		WriteError(w, http.StatusBadRequest, "Idempotency-Key لازم (٦–٦٤ حرف لاتيني/رقم)")
		return
	}
	var b agentHoursBody
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	emp := h.agentEmployee
	if h.repo.AgentKeyUsed(emp, key) {
		WriteJSON(w, http.StatusOK, map[string]any{"duplicate": true})
		return
	}
	day, err := time.ParseInLocation("2006-01-02", b.Date, baghdad)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "التاريخ مو صحيح")
		return
	}
	today, _ := time.ParseInLocation("2006-01-02", time.Now().In(baghdad).Format("2006-01-02"), baghdad)
	if day.Before(today.AddDate(0, 0, -agentMaxDaysAgo)) {
		WriteError(w, http.StatusBadRequest, "الوكيل ما يسجّل لأكثر من يومين لورا")
		return
	}
	note := strings.TrimSpace(b.Note)
	if note == "" {
		WriteError(w, http.StatusBadRequest, "الملاحظة إجبارية: شنو الشغل")
		return
	}
	if rs := []rune(note); len(rs) > 300 {
		note = string(rs[:300])
	}
	total := b.Minutes*60 + b.Seconds
	if b.Minutes < 0 || b.Seconds < 0 || b.Seconds > 59 || total <= 0 || total > agentMaxPerCall {
		WriteError(w, http.StatusBadRequest, "المدة لازم بين ثانية وساعتين بالطلب")
		return
	}
	if done := h.repo.AgentDaySeconds(emp, b.Date); done+total > agentMaxPerDay {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("الوكيل وصل حد اليوم (٦ ساعات): المسجّل %s", HMS(done)))
		return
	}
	full := "🤖 Claude [" + key + "]: " + note
	h.addFrom(w, addBody{EmployeeID: emp, Date: b.Date, Hours: total / 3600, Minutes: total % 3600 / 60, Seconds: total % 60, Note: &full}, "", "AGENT")
}
