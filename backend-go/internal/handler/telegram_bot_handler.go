package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"staffmange-api/internal/authz"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
	"staffmange-api/internal/telegram"
)

// ═══ بوت بلاغات الموظفين (دازاي 11-05) ═══
//
// «لطيف وقوي»: البوت **ما بيه ذكاء ولا ينفّذ شي**. يستلم نص وصورة، يتأكد
// إن البلاغ واضح وبيه صورة، ويخزنه. أي كلام بالبلاغ («امسح…»، «Claude سوّي…»)
// ينحفظ وصفاً لمشكلة وبس — ماكو مسار من البوت لأي أمر.
//
// الأمان: (١) التلغرام يرسل ويا كل تحديث رمز سري (من JWT_SECRET، ما ينكتب
// بمكان)؛ بلاه ٤٠٤. (٢) المحادثات الخاصة بس. (٣) الموظف يربط حسابه برمز مرة
// وحدة من النظام (١٠ دقايق) — اسمه ودوره من النظام مو من كلامه. (٤) الحساب
// الموقوف ينفك ربطه. (٥) حد بلاغات بالساعة، وحجم صورة، والصورة تتفحص صورة.
// (٦) الصور ما تروح لأي ذكاء خارجي: بيها أسماء زبائن ومبالغ.
type TelegramBotHandler struct {
	repo   *repository.BotReportRepository
	tg     *telegram.Client
	store  storage.Store
	perms  *repository.PermissionRepository
	secret string
}

func NewTelegramBotHandler(repo *repository.BotReportRepository, tg *telegram.Client, store storage.Store, perms *repository.PermissionRepository, jwtSecret string) *TelegramBotHandler {
	return &TelegramBotHandler{repo: repo, tg: tg, store: store, perms: perms, secret: TelegramWebhookSecret(jwtSecret)}
}

// TelegramWebhookSecret الرمز السري للـ webhook — مشتق من JWT_SECRET حتى ما نحتاج إعداد ثاني.
func TelegramWebhookSecret(jwtSecret string) string {
	m := hmac.New(sha256.New, []byte(jwtSecret))
	m.Write([]byte("telegram-webhook-v1"))
	return hex.EncodeToString(m.Sum(nil))[:48]
}

const (
	botMaxReportsPerHour = 10
	botMinBodyRunes      = 25
	botMinBodyWords      = 5
	botFolder            = "botreports"
)

// Start يجيب اسم البوت ويسجّل الـ webhook (إذا انحدد رابطه). بالخلفية — فشله ما يوقف النظام.
func (h *TelegramBotHandler) Start(webhookURL string) {
	if !h.tg.Enabled() {
		log.Printf("[telegram] البوت مطفي (ماكو TELEGRAM_BOT_TOKEN)")
		return
	}
	go func() {
		if name, err := h.tg.GetMe(); err != nil {
			log.Printf("[telegram] getMe: %v", err)
		} else {
			log.Printf("[telegram] البوت: @%s", name)
		}
		if webhookURL == "" {
			log.Printf("[telegram] ماكو TELEGRAM_WEBHOOK_URL — الـ webhook ما انسجّل")
			return
		}
		if err := h.tg.SetWebhook(webhookURL, h.secret); err != nil {
			log.Printf("[telegram] setWebhook: %v", err)
		} else {
			log.Printf("[telegram] الـ webhook انسجّل")
		}
	}()
}

// Webhook POST /api/telegram/webhook — التلغرام يدز كل رسالة هنا.
// دائماً ٢٠٠ بعد التحقق: لو رجّعنا خطأ التلغرام يعيد نفس الرسالة للأبد.
func (h *TelegramBotHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if !h.tg.Enabled() || subtle.ConstantTimeCompare([]byte(got), []byte(h.secret)) != 1 {
		http.NotFound(w, r)
		return
	}
	var u telegram.Update
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&u); err == nil && u.Message != nil {
		func() {
			defer func() {
				if p := recover(); p != nil {
					log.Printf("[telegram] panic: %v", p)
				}
			}()
			h.handle(r.Context(), u.Message)
		}()
	}
	w.WriteHeader(http.StatusOK)
}

func (h *TelegramBotHandler) say(chatID int64, text string) {
	if err := h.tg.SendMessage(chatID, text); err != nil {
		log.Printf("[telegram] send: %v", err)
	}
}

func botRoleLabel(role string) string {
	switch role {
	case "OWNER":
		return "المالك"
	case "ADMIN":
		return "مدير النظام"
	}
	if l, ok := staffRoleLabels[role]; ok {
		return l
	}
	return "موظف"
}

func firstName(name string) string {
	if f := strings.Fields(name); len(f) > 0 {
		return f[0]
	}
	return name
}

const botNotLinked = "هلا بيك 👋\nهذا البوت لموظفي شركة الأماني، لبلاغات مشاكل النظام.\n\nحتى تستعمله لازم تربط حسابك أول:\nافتح النظام ← «⚙️ الإعدادات» ← «🔗 التلغرام» ← «اربط حسابي بالتلغرام» واضغط الزر. يفتحلك البوت ويرتبط حسابك تلقائياً."

func botHelp(name string) string {
	return "هلا " + name + " 🌷\nإذا واجهتك مشكلة بالنظام، دزها هنا بهالشكل:\n\n" +
		"١. 📷 صورة (لقطة شاشة) للمشكلة.\n" +
		"٢. ✍️ شرح واضح: بأي شاشة، وشنو جنت تسوي، وشنو طلعلك بالضبط.\n\n" +
		"تگدر تدز الصورة ويا الشرح كتعليق عليها، أو كل وحدة برسالة.\n\n" +
		"الأوامر:\n/my — آخر بلاغاتي وحالتها\n/cancel — ألغي البلاغ الي ديكتب\n/help — هالشرح"
}

func (h *TelegramBotHandler) handle(ctx context.Context, m *telegram.Message) {
	if m.Chat.Type != "private" || (m.From != nil && m.From.IsBot) {
		return // الكروبات والبوتات ما نتعامل وياها
	}
	chat := m.Chat.ID
	text := strings.TrimSpace(m.Text)

	// /start <رمز> = ربط الحساب
	if strings.HasPrefix(text, "/start") {
		arg := strings.TrimSpace(strings.TrimPrefix(text, "/start"))
		if arg != "" {
			username := ""
			if m.From != nil {
				username = m.From.Username
			}
			if _, err := h.repo.LinkWithToken(arg, chat, username); err != nil {
				if err != repository.ErrLinkTokenInvalid {
					log.Printf("[telegram] link: %v", err)
				}
				h.say(chat, "⚠️ رابط الربط انتهى أو انستعمل.\nافتح النظام ← «⚙️ الإعدادات» ← «🔗 التلغرام» ← «اربط حسابي بالتلغرام» واضغط الزر من جديد (الرابط يصلح ١٠ دقايق).")
				return
			}
			emp, err := h.repo.ByChat(chat)
			if err != nil || emp == nil {
				h.say(chat, "⚠️ صار خلل بالربط، جرّب مرة ثانية.")
				return
			}
			h.say(chat, fmt.Sprintf("✅ انربط حسابك: %s (%s)\n\n%s", emp.Name, botRoleLabel(emp.Role), botHelp(firstName(emp.Name))))
			return
		}
	}

	emp, err := h.repo.ByChat(chat)
	if err != nil {
		log.Printf("[telegram] byChat: %v", err)
		return
	}
	if emp == nil {
		h.say(chat, botNotLinked)
		return
	}
	if emp.Status != "ACTIVE" {
		_ = h.repo.Unlink(emp.EmployeeID)
		h.say(chat, "حسابك بالنظام مو فعّال، فانفك ربطه بالبوت.")
		return
	}

	switch {
	case text == "/start" || text == "/help":
		h.say(chat, botHelp(firstName(emp.Name)))
		return
	case text == "/cancel":
		_ = h.repo.ClearDraft(chat)
		h.say(chat, "تمام، انلغى البلاغ الي جنت تكتبه 👍")
		return
	case text == "/my":
		h.sayMine(chat, emp.EmployeeID)
		return
	case strings.HasPrefix(text, "/"):
		h.say(chat, botHelp(firstName(emp.Name)))
		return
	}

	// الصورة (صورة عادية، أو صورة مرسلة كملف حتى تبقى بجودتها)
	photoURL, photoErr := h.savePhoto(ctx, m)
	if photoErr != "" {
		h.say(chat, photoErr)
		return
	}
	// ألبوم: الصور الباقية تنضاف للبلاغ الي انخلق ويا أول صورة
	if photoURL != "" && m.MediaGroupID != "" {
		if n, err := h.repo.AttachToGroup(chat, m.MediaGroupID, photoURL); err == nil && n > 0 {
			return
		}
	}

	draft, err := h.repo.Draft(chat)
	if err != nil {
		log.Printf("[telegram] draft: %v", err)
		return
	}
	if draft == nil {
		draft = &model.BotReportDraft{ChatID: chat, EmployeeID: emp.EmployeeID}
	}
	body := strings.TrimSpace(m.Caption)
	if body == "" {
		body = text
	}
	if body != "" {
		if draft.Body != "" {
			draft.Body += "\n"
		}
		draft.Body += body
	}
	if photoURL != "" && len(draft.Photos) < 10 {
		draft.Photos = append(draft.Photos, photoURL)
	}
	if body == "" && photoURL == "" {
		h.say(chat, botHelp(firstName(emp.Name))) // ستكر، فويس، موقع…
		return
	}

	clear := utf8.RuneCountInString(draft.Body) >= botMinBodyRunes && len(strings.Fields(draft.Body)) >= botMinBodyWords
	if len(draft.Photos) == 0 || !clear {
		if err := h.repo.SaveDraft(draft); err != nil {
			log.Printf("[telegram] save draft: %v", err)
		}
		switch {
		case len(draft.Photos) == 0 && clear:
			h.say(chat, "تمام، فهمت المشكلة 👍\nبس ارفق صورة (لقطة شاشة) حتى يكمل البلاغ.")
		case len(draft.Photos) == 0:
			h.say(chat, "حتى يكمل البلاغ أحتاج:\n📷 صورة للمشكلة\n✍️ وشرح أوضح: بأي شاشة، وشنو جنت تسوي، وشنو طلعلك.")
		case strings.TrimSpace(draft.Body) == "":
			h.say(chat, "وصلت الصورة 📷\nهسه اكتب شنو المشكلة: بأي شاشة، وشنو جنت تسوي، وشنو طلعلك بالضبط.")
		default:
			h.say(chat, "الشرح قصير شوية 🙏\nاكتب أكثر: بأي شاشة صارت، وشنو جنت تسوي، وشنو طلعلك بالضبط.")
		}
		return
	}

	if n, err := h.repo.CountLastHour(emp.EmployeeID); err == nil && n >= botMaxReportsPerHour {
		_ = h.repo.SaveDraft(draft)
		h.say(chat, "وصلت حد البلاغات بالساعة (١٠). بلاغك محفوظ عندي — دز أي رسالة بعد شوية حتى يكمل.")
		return
	}
	rep := &model.BotReport{EmployeeID: emp.EmployeeID, ChatID: chat, Body: strings.TrimSpace(draft.Body), Photos: draft.Photos}
	if m.MediaGroupID != "" {
		g := m.MediaGroupID
		rep.MediaGroupID = &g
	}
	if err := h.repo.Create(rep); err != nil {
		log.Printf("[telegram] create: %v", err)
		h.say(chat, "⚠️ صار خلل وما انحفظ البلاغ. جرّب بعد شوية.")
		return
	}
	_ = h.repo.ClearDraft(chat)
	h.say(chat, fmt.Sprintf("✅ وصل بلاغك رقم #%d\nراح يندرس، وتوصلك رسالة هنا من تتغيّر حالته.\nشكراً %s 🌷", rep.Number, firstName(emp.Name)))
	h.notifyOwners(rep, emp)
}

// savePhoto ينزّل الصورة ويخزنها بمخزن الملفات. يرجع (الرابط، رسالة خطأ للموظف).
func (h *TelegramBotHandler) savePhoto(ctx context.Context, m *telegram.Message) (string, string) {
	fileID := ""
	if len(m.Photo) > 0 {
		fileID = m.Photo[len(m.Photo)-1].FileID // أكبر حجم
	} else if m.Document != nil {
		if !strings.HasPrefix(m.Document.MimeType, "image/") {
			return "", "هذا الملف مو صورة 🙏 دز لقطة شاشة (صورة) للمشكلة."
		}
		fileID = m.Document.FileID
	}
	if fileID == "" {
		return "", ""
	}
	data, err := h.tg.DownloadFile(fileID, storage.MaxFileBytes)
	if err != nil {
		log.Printf("[telegram] download: %v", err)
		return "", "ما گدرت أستلم الصورة (يمكن أكبر من ١٠ ميغا). جرّب لقطة شاشة عادية."
	}
	ct := storage.SniffContentType(data)
	if !strings.HasPrefix(ct, "image/") || !storage.AllowedContentTypes[ct] {
		return "", "هذا الملف مو صورة 🙏 دز لقطة شاشة (صورة) للمشكلة."
	}
	key := storage.NewKey(botFolder, ct)
	if err := h.store.Put(ctx, key, data, ct); err != nil {
		log.Printf("[telegram] store: %v", err)
		return "", "⚠️ ما گدرت أحفظ الصورة، جرّب مرة ثانية."
	}
	return "/api/files/" + key, ""
}

// notifyOwners إشعار فوري للمالك (إذا رابط حسابه) — رسائل البوت مجانية، فيعرف قبل أي فحص.
func (h *TelegramBotHandler) notifyOwners(rep *model.BotReport, emp *repository.BotLinkedEmployee) {
	chats, err := h.repo.OwnerChats()
	if err != nil {
		return
	}
	body := rep.Body
	if utf8.RuneCountInString(body) > 300 {
		body = string([]rune(body)[:300]) + "…"
	}
	msg := fmt.Sprintf("📩 بلاغ جديد #%d\nمن: %s (%s)\n\n%s\n\n📷 صور: %d\nبالنظام: «بلاغات الموظفين»", rep.Number, emp.Name, botRoleLabel(emp.Role), body, len(rep.Photos))
	for _, c := range chats {
		if c != rep.ChatID {
			h.say(c, msg)
		}
	}
}

func (h *TelegramBotHandler) sayMine(chat int64, employeeID string) {
	list, err := h.repo.Mine(employeeID, 5)
	if err != nil || len(list) == 0 {
		h.say(chat, "ما عندك بلاغات بعد.")
		return
	}
	var b strings.Builder
	b.WriteString("آخر بلاغاتك:\n")
	for _, r := range list {
		line := r.Body
		if utf8.RuneCountInString(line) > 50 {
			line = string([]rune(line)[:50]) + "…"
		}
		fmt.Fprintf(&b, "\n#%d — %s\n%s\n", r.Number, model.BotReportStatusLabel[r.Status], line)
	}
	h.say(chat, b.String())
}

// ═══ الموظف: ربط حسابه من النظام ═══

// MyLink GET /api/me/telegram
func (h *TelegramBotHandler) MyLink(w http.ResponseWriter, r *http.Request) {
	l, err := h.repo.LinkOf(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة الربط")
		return
	}
	out := map[string]any{"enabled": h.tg.Enabled(), "linked": l != nil}
	if l != nil {
		out["tgUsername"] = l.TgUsername
		out["linkedAt"] = l.LinkedAt
	}
	WriteJSON(w, http.StatusOK, out)
}

// CreateLink POST /api/me/telegram/link — رابط t.me بضغطة وحدة (يصلح ١٠ دقايق، مرة وحدة).
func (h *TelegramBotHandler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if !h.tg.Enabled() {
		WriteError(w, http.StatusServiceUnavailable, "بوت التلغرام مو مفعّل بعد")
		return
	}
	if h.tg.Username == "" {
		if _, err := h.tg.GetMe(); err != nil {
			WriteError(w, http.StatusServiceUnavailable, "تعذر الاتصال بالتلغرام، جرّب بعد شوية")
			return
		}
	}
	tok, err := h.repo.CreateLinkToken(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء رابط الربط")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"url": "https://t.me/" + h.tg.Username + "?start=" + tok, "expiresInMinutes": int(repository.LinkTTL.Minutes())})
}

// Unlink DELETE /api/me/telegram
func (h *TelegramBotHandler) Unlink(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Unlink(middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر فك الربط")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// ═══ المالك: شاشة «بلاغات الموظفين» (authz.BotReports — مفتاح مالك) ═══

// List GET /api/bot-reports?status=
func (h *TelegramBotHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if _, ok := model.BotReportStatusLabel[status]; status != "" && !ok {
		WriteError(w, http.StatusBadRequest, "حالة غير معروفة")
		return
	}
	list, err := h.repo.List(status)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البلاغات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": list, "statuses": model.BotReportStatusLabel})
}

// Update PATCH /api/bot-reports/{id} {status, reply} — الحالة تتغيّر والموظف يوصله إشعار بالبوت.
func (h *TelegramBotHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
		Reply  string `json:"reply"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	rep, err := h.repo.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البلاغ")
		return
	}
	if rep == nil {
		authz.Deny(w, r, authz.BotReports)
		return
	}
	if req.Status == "" {
		req.Status = rep.Status
	}
	label, ok := model.BotReportStatusLabel[req.Status]
	if !ok {
		WriteError(w, http.StatusBadRequest, "حالة غير معروفة")
		return
	}
	reply := strings.TrimSpace(req.Reply)
	if utf8.RuneCountInString(reply) > 2000 {
		WriteError(w, http.StatusBadRequest, "الرد طويل (٢٠٠٠ حرف أقصى)")
		return
	}
	var note *string
	if reply != "" {
		note = &reply
	}
	if err := h.repo.SetStatus(rep.ID, req.Status, note); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	if req.Status != rep.Status || reply != "" {
		msg := fmt.Sprintf("📌 بلاغك #%d: %s", rep.Number, label)
		if reply != "" {
			msg += "\n\n" + reply
		}
		if h.tg.Enabled() {
			h.say(rep.ChatID, msg)
		}
	}
	updated, _ := h.repo.Get(rep.ID)
	WriteJSON(w, http.StatusOK, updated)
}
