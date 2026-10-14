package handler

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"

	"staffmange-api/internal/mail"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// ═══ تطبيق الزبائن — قرار (ع) 10-14 ═══
//
// واجهة مستقلة للزبون: يسوي حسابه بنفسه (اسم مستخدم + باسورد، والإيميل
// للتأكيد برمز OTP بس)، يشوف خدماتنا وعروضنا، يحچي ويّا ماتركس الي يستلم
// حجزه، ويشوف حجوزاته.
//
// 🔴 العزل: توكن الزبون موقّع بمفتاح مشتق مختلف عن توكن الموظفين وبطبقة
// "customer" — فما يفتح أي مسار موظفين (RequireAuth يرفض أي طبقة غير
// staff)، وتوكن الموظف ما يفتح مسارات الزبون.

const (
	customerRealm      = "customer"
	customerTokenTTL   = 30 * 24 * time.Hour
	otpTTL             = 10 * time.Minute
	otpMaxAttempts     = 5
	otpResendGap       = 60 * time.Second
	customerDailyChats = 60
)

var usernameRe = regexp.MustCompile(`^[a-z0-9_.]{3,30}$`)

type CustomerAppHandler struct {
	db        *sqlx.DB
	key       []byte // مفتاح توقيع توكن الزبون (مشتق)
	otpKey    []byte
	mail      *mail.Sender
	matrix    *service.CustomerMatrix
	store     storage.Store
	bookings  *service.BookingService
	customers *repository.CustomerRepository
	notify    *repository.NotificationRepository
}

func NewCustomerAppHandler(db *sqlx.DB, jwtSecret string, m *mail.Sender, matrix *service.CustomerMatrix, store storage.Store,
	bookings *service.BookingService, customers *repository.CustomerRepository, notify *repository.NotificationRepository) *CustomerAppHandler {
	derive := func(label string) []byte {
		mac := hmac.New(sha256.New, []byte(jwtSecret))
		mac.Write([]byte(label))
		return mac.Sum(nil)
	}
	h := &CustomerAppHandler{db: db, key: derive("customer-app-token"), otpKey: derive("customer-app-otp"), mail: m,
		matrix: matrix, store: store, bookings: bookings, customers: customers, notify: notify}
	if matrix != nil {
		matrix.CreateBooking = h.createBookingFromMatrix
	}
	return h
}

type customerAccount struct {
	ID                    string     `db:"id" json:"id"`
	Username              string     `db:"username" json:"username"`
	Email                 string     `db:"email" json:"email"`
	PasswordHash          string     `db:"passwordHash" json:"-"`
	Name                  string     `db:"name" json:"name"`
	Phone                 string     `db:"phone" json:"phone"`
	City                  *string    `db:"city" json:"city"`
	CustomerID            *string    `db:"customerId" json:"-"`
	EmailVerifiedAt       *time.Time `db:"emailVerifiedAt" json:"-"`
	SessionsInvalidatedAt *time.Time `db:"sessionsInvalidatedAt" json:"-"`
	LastSeenPromoAt       *time.Time `db:"lastSeenPromoAt" json:"-"`
	CreatedAt             time.Time  `db:"createdAt" json:"createdAt"`
}

func (h *CustomerAppHandler) account(where string, arg any) (*customerAccount, error) {
	var a customerAccount
	if err := h.db.Get(&a, `SELECT * FROM "CustomerAccount" WHERE `+where, arg); err != nil {
		return nil, err
	}
	return &a, nil
}

func maskEmail(e string) string {
	at := strings.Index(e, "@")
	if at <= 1 {
		return e
	}
	return e[:1] + strings.Repeat("*", min(at-1, 5)) + e[at:]
}

// ─── الرموز ───

func (h *CustomerAppHandler) hashCode(accountID, code string) string {
	mac := hmac.New(sha256.New, h.otpKey)
	mac.Write([]byte(accountID + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *CustomerAppHandler) sendOtp(a *customerAccount, purpose string) error {
	var last time.Time
	if err := h.db.Get(&last, `SELECT "createdAt" FROM "CustomerOtp" WHERE "accountId" = $1 ORDER BY "createdAt" DESC LIMIT 1`, a.ID); err == nil &&
		time.Since(last) < otpResendGap {
		return errTooSoon
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	if _, err := h.db.Exec(`UPDATE "CustomerOtp" SET "usedAt" = now() WHERE "accountId" = $1 AND purpose = $2 AND "usedAt" IS NULL`, a.ID, purpose); err != nil {
		return err
	}
	if _, err := h.db.Exec(`INSERT INTO "CustomerOtp" (id, "accountId", purpose, "codeHash", "expiresAt") VALUES ($1,$2,$3,$4,$5)`,
		uuid.NewString(), a.ID, purpose, h.hashCode(a.ID, code), time.Now().Add(otpTTL)); err != nil {
		return err
	}
	subject := "رمز تأكيد حسابك — الأماني"
	if purpose == "reset" {
		subject = "رمز تغيير كلمة المرور — الأماني"
	}
	return h.mail.Send(a.Email, subject, mail.OtpHTML(a.Name, code, purpose))
}

var errTooSoon = errors.New("انتظر دقيقة قبل ما تطلب رمز جديد")

// checkOtp يتحقق من الرمز ويستهلكه.
func (h *CustomerAppHandler) checkOtp(accountID, purpose, code string) error {
	var o struct {
		ID        string    `db:"id"`
		CodeHash  string    `db:"codeHash"`
		Attempts  int       `db:"attempts"`
		ExpiresAt time.Time `db:"expiresAt"`
	}
	err := h.db.Get(&o, `SELECT id, "codeHash", attempts, "expiresAt" FROM "CustomerOtp"
		WHERE "accountId" = $1 AND purpose = $2 AND "usedAt" IS NULL ORDER BY "createdAt" DESC LIMIT 1`, accountID, purpose)
	if err != nil {
		return errors.New("ماكو رمز فعّال — اطلب رمز جديد")
	}
	if time.Now().After(o.ExpiresAt) {
		return errors.New("الرمز انتهت صلاحيته — اطلب رمز جديد")
	}
	if o.Attempts >= otpMaxAttempts {
		return errors.New("محاولات كثيرة — اطلب رمز جديد")
	}
	if !hmac.Equal([]byte(h.hashCode(accountID, strings.TrimSpace(code))), []byte(o.CodeHash)) {
		_, _ = h.db.Exec(`UPDATE "CustomerOtp" SET attempts = attempts + 1 WHERE id = $1`, o.ID)
		return errors.New("الرمز غلط")
	}
	_, err = h.db.Exec(`UPDATE "CustomerOtp" SET "usedAt" = now() WHERE id = $1`, o.ID)
	return err
}

// ─── التوكن ───

type customerClaims struct {
	AccountID string `json:"cid"`
	Realm     string `json:"realm"`
	jwt.RegisteredClaims
}

func (h *CustomerAppHandler) token(id string) (string, error) {
	c := customerClaims{AccountID: id, Realm: customerRealm, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(customerTokenTTL)), IssuedAt: jwt.NewNumericDate(time.Now()),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(h.key)
}

type customerCtxKey struct{}

// RequireCustomer حارس مسارات الزبون.
func (h *CustomerAppHandler) RequireCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		c := &customerClaims{}
		t, err := jwt.ParseWithClaims(raw, c, func(*jwt.Token) (any, error) { return h.key, nil }, jwt.WithValidMethods([]string{"HS256"}))
		if raw == "" || err != nil || !t.Valid || c.Realm != customerRealm || c.AccountID == "" {
			WriteError(w, http.StatusUnauthorized, "سجّل دخولك")
			return
		}
		a, err := h.account("id = $1", c.AccountID)
		if err != nil || a.EmailVerifiedAt == nil {
			WriteError(w, http.StatusUnauthorized, "سجّل دخولك")
			return
		}
		if a.SessionsInvalidatedAt != nil && c.IssuedAt != nil && c.IssuedAt.Time.Add(time.Second).Before(*a.SessionsInvalidatedAt) {
			WriteError(w, http.StatusUnauthorized, "انتهت الجلسة — سجّل دخولك")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), customerCtxKey{}, a)))
	})
}

func current(r *http.Request) *customerAccount {
	a, _ := r.Context().Value(customerCtxKey{}).(*customerAccount)
	return a
}

func (h *CustomerAppHandler) session(w http.ResponseWriter, a *customerAccount) {
	tok, err := h.token(a.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء الجلسة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"token": tok, "account": a})
}

// ─── الحساب ───

// POST /api/c/signup
func (h *CustomerAppHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var b struct{ Name, Username, Email, Phone, Password string }
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	b.Name, b.Phone = strings.TrimSpace(b.Name), strings.TrimSpace(b.Phone)
	b.Username = strings.ToLower(strings.TrimSpace(b.Username))
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
	switch {
	case len([]rune(b.Name)) < 2:
		WriteError(w, http.StatusBadRequest, "اكتب اسمك")
		return
	case !usernameRe.MatchString(b.Username):
		WriteError(w, http.StatusBadRequest, "اسم المستخدم من 3 إلى 30 حرف إنكليزي صغير أو رقم (ويسمح _ و .)")
		return
	case !strings.Contains(b.Email, "@") || !strings.Contains(b.Email[strings.Index(b.Email, "@"):], ".") || len(b.Email) > 120:
		WriteError(w, http.StatusBadRequest, "الإيميل غير صحيح")
		return
	case len(b.Phone) < 10:
		WriteError(w, http.StatusBadRequest, "رقم الهاتف غير صحيح")
		return
	}
	if err := service.ValidatePasswordStrength(b.Password); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "خطأ")
		return
	}
	// حساب ما تأكد بعد بنفس الاسم والإيميل: نحدّثه ونعيد الرمز (حتى ما ينقفل الاسم)
	var existing []customerAccount
	_ = h.db.Select(&existing, `SELECT * FROM "CustomerAccount" WHERE username = $1 OR email = $2`, b.Username, b.Email)
	var a *customerAccount
	for i := range existing {
		e := &existing[i]
		if e.EmailVerifiedAt == nil && e.Username == b.Username && e.Email == b.Email {
			a = e
			continue
		}
		if e.Username == b.Username {
			WriteError(w, http.StatusConflict, "اسم المستخدم مأخوذ — جرّب اسم ثاني")
			return
		}
		WriteError(w, http.StatusConflict, "هذا الإيميل مسجّل بحساب ثاني")
		return
	}
	if a != nil {
		_, err = h.db.Exec(`UPDATE "CustomerAccount" SET name=$2, phone=$3, "passwordHash"=$4 WHERE id=$1`, a.ID, b.Name, b.Phone, string(hash))
		a.Name, a.Phone = b.Name, b.Phone
	} else {
		a = &customerAccount{ID: uuid.NewString(), Username: b.Username, Email: b.Email, Name: b.Name, Phone: b.Phone}
		_, err = h.db.Exec(`INSERT INTO "CustomerAccount" (id, username, email, "passwordHash", name, phone) VALUES ($1,$2,$3,$4,$5,$6)`,
			a.ID, a.Username, a.Email, string(hash), a.Name, a.Phone)
	}
	if err != nil {
		WriteError(w, http.StatusConflict, "تعذر إنشاء الحساب — يمكن الاسم أو الإيميل مستخدم")
		return
	}
	if err := h.sendOtp(a, "signup"); err != nil && !errors.Is(err, errTooSoon) {
		log.Printf("customer otp: %v", err)
		WriteError(w, http.StatusBadGateway, "ما كدرنا نرسل الرمز للإيميل — تأكد منه وحاول مرة ثانية")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"username": a.Username, "email": maskEmail(a.Email)})
}

// POST /api/c/verify {username, code}
func (h *CustomerAppHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var b struct{ Username, Code string }
	_ = DecodeJSON(r, &b)
	a, err := h.account("username = $1", strings.ToLower(strings.TrimSpace(b.Username)))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "الحساب غير موجود")
		return
	}
	if a.EmailVerifiedAt == nil {
		if err := h.checkOtp(a.ID, "signup", b.Code); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET "emailVerifiedAt" = now() WHERE id = $1`, a.ID)
		now := time.Now()
		a.EmailVerifiedAt = &now
		h.linkCustomer(a)
	}
	h.session(w, a)
}

// linkCustomer يربط الحساب بسجل الزبون الموجود إذا هاتفه مسجّل عدنا.
func (h *CustomerAppHandler) linkCustomer(a *customerAccount) {
	if a.CustomerID != nil {
		return
	}
	if c, err := h.customers.FindByPhone(a.Phone); err == nil && c != nil {
		_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET "customerId" = $2 WHERE id = $1`, a.ID, c.ID)
		a.CustomerID = &c.ID
	}
}

// POST /api/c/resend {username, purpose}
func (h *CustomerAppHandler) Resend(w http.ResponseWriter, r *http.Request) {
	var b struct{ Username, Purpose string }
	_ = DecodeJSON(r, &b)
	a, err := h.account("username = $1", strings.ToLower(strings.TrimSpace(b.Username)))
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
		return
	}
	purpose := "signup"
	if b.Purpose == "reset" {
		purpose = "reset"
	} else if a.EmailVerifiedAt != nil {
		WriteError(w, http.StatusBadRequest, "الحساب متأكد — سجّل دخولك")
		return
	}
	if err := h.sendOtp(a, purpose); err != nil {
		code := http.StatusBadGateway
		if errors.Is(err, errTooSoon) {
			code = http.StatusTooManyRequests
		}
		WriteError(w, code, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "email": maskEmail(a.Email)})
}

// POST /api/c/login {username, password}
func (h *CustomerAppHandler) Login(w http.ResponseWriter, r *http.Request) {
	var b struct{ Username, Password string }
	_ = DecodeJSON(r, &b)
	a, err := h.account("username = $1", strings.ToLower(strings.TrimSpace(b.Username)))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(b.Password)) != nil {
		WriteError(w, http.StatusUnauthorized, "اسم المستخدم أو كلمة المرور غلط")
		return
	}
	if a.EmailVerifiedAt == nil {
		_ = h.sendOtp(a, "signup")
		WriteJSON(w, http.StatusForbidden, map[string]any{"error": "لازم تأكد إيميلك أول — دزّينا رمز جديد", "needVerify": true, "email": maskEmail(a.Email)})
		return
	}
	h.linkCustomer(a)
	h.session(w, a)
}

// POST /api/c/forgot {username}
func (h *CustomerAppHandler) Forgot(w http.ResponseWriter, r *http.Request) {
	var b struct{ Username string }
	_ = DecodeJSON(r, &b)
	a, err := h.account("username = $1", strings.ToLower(strings.TrimSpace(b.Username)))
	if err != nil || a.EmailVerifiedAt == nil {
		WriteError(w, http.StatusBadRequest, "ما لگينا حساب متأكد بهذا الاسم")
		return
	}
	if err := h.sendOtp(a, "reset"); err != nil && !errors.Is(err, errTooSoon) {
		WriteError(w, http.StatusBadGateway, "ما كدرنا نرسل الرمز — حاول بعد شوية")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "email": maskEmail(a.Email)})
}

// POST /api/c/reset {username, code, password}
func (h *CustomerAppHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var b struct{ Username, Code, Password string }
	_ = DecodeJSON(r, &b)
	if err := service.ValidatePasswordStrength(b.Password); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	a, err := h.account("username = $1", strings.ToLower(strings.TrimSpace(b.Username)))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "الحساب غير موجود")
		return
	}
	if err := h.checkOtp(a.ID, "reset", b.Code); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET "passwordHash" = $2, "sessionsInvalidatedAt" = now() - interval '2 seconds' WHERE id = $1`, a.ID, string(hash))
	h.session(w, a)
}

// GET /api/c/me
func (h *CustomerAppHandler) Me(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, current(r))
}

// PUT /api/c/me {name, phone}
func (h *CustomerAppHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	a := current(r)
	var b struct{ Name, Phone, City string }
	_ = DecodeJSON(r, &b)
	if c := strings.TrimSpace(b.City); c != "" {
		a.City = &c
	}
	if n := strings.TrimSpace(b.Name); len([]rune(n)) >= 2 {
		a.Name = n
	}
	if p := strings.TrimSpace(b.Phone); len(p) >= 10 {
		a.Phone = p
	}
	_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET name = $2, phone = $3, city = $4 WHERE id = $1`, a.ID, a.Name, a.Phone, a.City)
	h.linkCustomer(a)
	WriteJSON(w, http.StatusOK, a)
}

// ─── المحتوى ───

type promoItem struct {
	ID        string    `db:"id" json:"id"`
	Kind      string    `db:"kind" json:"kind"`
	Title     string    `db:"title" json:"title"`
	Body      *string   `db:"body" json:"body"`
	ImageURL  *string   `db:"imageUrl" json:"-"`
	HasImage  bool      `db:"-" json:"hasImage"`
	ServiceID *string   `db:"serviceId" json:"serviceId"`
	Active    bool      `db:"active" json:"active"`
	Ord       int       `db:"ord" json:"ord"`
	Discount  *int      `db:"discountPercent" json:"discountPercent"`
	Source    string    `db:"source" json:"source"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	IsNew     bool      `db:"-" json:"isNew"`
}

func (h *CustomerAppHandler) promos(onlyActive bool) []promoItem {
	q := `SELECT id, kind, title, body, "imageUrl", "serviceId", active, ord, "discountPercent", source, "createdAt" FROM "PromoItem"`
	if onlyActive {
		q += ` WHERE active`
	}
	rows := []promoItem{}
	_ = h.db.Select(&rows, q+` ORDER BY ord, "createdAt" DESC`)
	for i := range rows {
		rows[i].HasImage = rows[i].ImageURL != nil && *rows[i].ImageURL != ""
	}
	return rows
}

// GET /api/c/home — العروض والأعمال والخدمات.
func (h *CustomerAppHandler) Home(w http.ResponseWriter, r *http.Request) {
	a := current(r)
	items := h.promos(true)
	newCount := 0
	for i := range items {
		if a.LastSeenPromoAt == nil || items[i].CreatedAt.After(*a.LastSeenPromoAt) {
			items[i].IsNew = true
			newCount++
		}
	}
	services := []struct {
		ID       string  `db:"id" json:"id"`
		Name     string  `db:"name" json:"name"`
		Category *string `db:"category" json:"category"`
	}{}
	_ = h.db.Select(&services, `SELECT id, name, category FROM "Service" WHERE category IS DISTINCT FROM 'إدارية' ORDER BY category NULLS LAST, name`)
	var done int
	_ = h.db.Get(&done, `SELECT count(*) FROM "Booking" WHERE status = 'COMPLETED'`)
	WriteJSON(w, http.StatusOK, map[string]any{"items": items, "services": services, "newCount": newCount, "completedJobs": done,
		"matrixOn": h.matrix.Enabled()})
}

// POST /api/c/home/seen — شاف العروض الجديدة.
func (h *CustomerAppHandler) Seen(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET "lastSeenPromoAt" = now() WHERE id = $1`, current(r).ID)
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/c/img/{id} — صورة عرض/عمل (محتوى ترويجي عام).
func (h *CustomerAppHandler) Image(w http.ResponseWriter, r *http.Request) {
	var key string
	if err := h.db.Get(&key, `SELECT COALESCE("imageUrl", '') FROM "PromoItem" WHERE id = $1`, r.PathValue("id")); err != nil || key == "" {
		http.NotFound(w, r)
		return
	}
	data, ct, err := h.store.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

// GET /api/c/bookings
func (h *CustomerAppHandler) Bookings(w http.ResponseWriter, r *http.Request) {
	rows, err := service.CustomerBookings(h.db, current(r).ID, 100)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// ─── ماتركس ───

type custChatMsg struct {
	ID        string    `db:"id" json:"id"`
	Role      string    `db:"role" json:"role"`
	Content   string    `db:"content" json:"content"`
	BookingID *string   `db:"bookingId" json:"bookingCode,omitempty"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

func (h *CustomerAppHandler) messages(accountID string, limit int) []custChatMsg {
	rows := []custChatMsg{}
	_ = h.db.Select(&rows, `SELECT * FROM (SELECT id, role, content, "bookingId", "createdAt" FROM "CustomerChatMessage"
		WHERE "accountId" = $1 ORDER BY "createdAt" DESC LIMIT $2) x ORDER BY "createdAt"`, accountID, limit)
	return rows
}

// GET /api/c/chat
func (h *CustomerAppHandler) Chat(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"messages": h.messages(current(r).ID, 100), "matrixOn": h.matrix.Enabled()})
}

// POST /api/c/chat {text}
func (h *CustomerAppHandler) Send(w http.ResponseWriter, r *http.Request) {
	a := current(r)
	var b struct{ Text string }
	_ = DecodeJSON(r, &b)
	text := strings.TrimSpace(b.Text)
	if text == "" || len([]rune(text)) > 1500 {
		WriteError(w, http.StatusBadRequest, "اكتب رسالتك (أقصى 1500 حرف)")
		return
	}
	var today int
	_ = h.db.Get(&today, `SELECT count(*) FROM "CustomerChatMessage" WHERE "accountId" = $1 AND role = 'user' AND "createdAt" > now() - interval '24 hours'`, a.ID)
	if today >= customerDailyChats {
		WriteError(w, http.StatusTooManyRequests, "وصلت حد الرسائل لليوم — فريقنا راح يتواصل وياك، أو ارجعلنا باچر 🌙")
		return
	}
	history := []service.AgentTurn{}
	for _, m := range h.messages(a.ID, 20) {
		role := "USER"
		if m.Role == "matrix" {
			role = "ASSISTANT"
		}
		history = append(history, service.AgentTurn{Role: role, Text: m.Content})
	}
	_, _ = h.db.Exec(`INSERT INTO "CustomerChatMessage" (id, "accountId", role, content) VALUES ($1,$2,'user',$3)`, uuid.NewString(), a.ID, text)
	answer, code, err := h.matrix.Chat(a.ID, a.Name, history, text)
	if err != nil {
		log.Printf("customer matrix: %v", err)
		if answer == "" {
			answer = "عذراً، صار عندي ضغط هسه 🙏 حاول بعد شوية، أو إذا مستعجل اكتبلي رقمك وفريقنا يتواصل وياك."
		}
	}
	if code != "" && answer == "" {
		answer = "تم تسجيل حجزك ✅ رقم الحجز " + code + " — فريقنا راح يتواصل وياك لتثبيت الموعد."
	}
	var bid *string
	if code != "" {
		bid = &code
	}
	m := custChatMsg{ID: uuid.NewString(), Role: "matrix", Content: answer, BookingID: bid, CreatedAt: time.Now()}
	_, _ = h.db.Exec(`INSERT INTO "CustomerChatMessage" (id, "accountId", role, content, "bookingId") VALUES ($1,$2,'matrix',$3,$4)`, m.ID, a.ID, m.Content, bid)
	WriteJSON(w, http.StatusOK, m)
}

// createBookingFromMatrix ماتركس يرحّل طلب الزبون للحجوزات.
func (h *CustomerAppHandler) createBookingFromMatrix(accountID string, d service.CustomerBookingDraft) (string, error) {
	a, err := h.account("id = $1", accountID)
	if err != nil {
		return "", err
	}
	if len(d.ServiceIDs) == 0 {
		return "", errors.New("حدد الخدمة")
	}
	if strings.TrimSpace(d.Address) == "" {
		return "", errors.New("العنوان مطلوب")
	}
	// ماتركس يحجز بس لنفس الزبون — وأقصى 3 حجوزات مفتوحة من التطبيق
	var open int
	_ = h.db.Get(&open, `SELECT count(*) FROM "Booking" WHERE "customerAccountId" = $1 AND status = 'PENDING'`, accountID)
	if open >= 3 {
		return "", errors.New("عندك 3 طلبات بانتظار التثبيت — فريقنا راح يتواصل وياك عليهن أول")
	}
	if a.CustomerID == nil {
		h.linkCustomer(a)
	}
	if a.CustomerID == nil {
		c, err := h.customers.Create(a.Name, a.Phone, &d.Address, nil, nil)
		if err != nil {
			return "", err
		}
		_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET "customerId" = $2 WHERE id = $1`, a.ID, c.ID)
		a.CustomerID = &c.ID
	}
	notes := "🤖 طلب من تطبيق الزبائن عن طريق ماتركس"
	if d.PreferredTime != "" {
		notes += "\nالوقت المناسب للزبون: " + d.PreferredTime
	}
	if d.Notes != "" {
		notes += "\n" + d.Notes
	}
	req := model.CreateBookingRequest{CustomerID: *a.CustomerID, ServiceIDs: d.ServiceIDs, Notes: &notes, Address: &d.Address}
	req.ServiceID = &d.ServiceIDs[0]
	if d.LocationURL != "" {
		req.LocationUrl = &d.LocationURL
	}
	if d.DeviceCount > 0 {
		req.DeviceCount = &d.DeviceCount
	}
	b, err := h.bookings.Create(req)
	if err != nil {
		return "", err
	}
	_, _ = h.db.Exec(`UPDATE "Booking" SET source = 'matrix_app', "customerAccountId" = $2 WHERE id = $1`, b.ID, a.ID)
	_ = h.notify.CreateForRolesOrPermission([]string{"ADMIN", "MONITOR"}, "sales_booking", "BOOKING",
		"🤖 ماتركس سجّل حجز جديد من تطبيق الزبائن: "+b.Code+" — "+a.Name)
	return b.Code, nil
}

// ═══ إدارة المالك ═══

// GET /api/customer-app/promos
func (h *CustomerAppHandler) AdminPromos(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.promos(false))
}

type promoInput struct {
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	Body      *string `json:"body"`
	ServiceID *string `json:"serviceId"`
	Active    bool    `json:"active"`
	Ord       int     `json:"ord"`
	Discount  *int    `json:"discountPercent"`
}

func (p *promoInput) valid() error {
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return errors.New("العنوان مطلوب")
	}
	if p.Kind != "offer" && p.Kind != "work" && p.Kind != "service" {
		p.Kind = "offer"
	}
	if p.ServiceID != nil && *p.ServiceID == "" {
		p.ServiceID = nil
	}
	return nil
}

// POST /api/customer-app/promos
func (h *CustomerAppHandler) AdminCreatePromo(w http.ResponseWriter, r *http.Request) {
	var p promoInput
	_ = DecodeJSON(r, &p)
	if err := p.valid(); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	id := uuid.NewString()
	if _, err := h.db.Exec(`INSERT INTO "PromoItem" (id, kind, title, body, "serviceId", active, ord, "discountPercent", "approvedAt") VALUES ($1,$2,$3,$4,$5,$6,$7,$8, CASE WHEN $6 THEN now() END)`,
		id, p.Kind, p.Title, p.Body, p.ServiceID, p.Active, p.Ord, p.Discount); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id})
}

// PUT /api/customer-app/promos/{id}
func (h *CustomerAppHandler) AdminUpdatePromo(w http.ResponseWriter, r *http.Request) {
	var p promoInput
	_ = DecodeJSON(r, &p)
	if err := p.valid(); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, err := h.db.Exec(`UPDATE "PromoItem" SET kind=$2, title=$3, body=$4, "serviceId"=$5, active=$6, ord=$7,
		"discountPercent"=$8,
		"approvedAt" = CASE WHEN $6 THEN COALESCE("approvedAt", now()) ELSE "approvedAt" END WHERE id=$1`,
		r.PathValue("id"), p.Kind, p.Title, p.Body, p.ServiceID, p.Active, p.Ord, p.Discount)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /api/customer-app/promos/{id}
func (h *CustomerAppHandler) AdminDeletePromo(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "PromoItem" WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/customer-app/promos/{id}/image (multipart: file)
func (h *CustomerAppHandler) AdminPromoImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر قراءة الصورة")
		return
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو صورة")
		return
	}
	defer f.Close()
	data, err := storage.ReadLimited(f, storage.MaxFileBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	ct := storage.SniffContentType(data)
	if !strings.HasPrefix(ct, "image/") || !storage.AllowedContentTypes[ct] {
		WriteError(w, http.StatusBadRequest, "لازم صورة JPG/PNG/WEBP")
		return
	}
	key := storage.NewKey("customer-app", ct)
	if err := h.store.Put(r.Context(), key, data, ct); err != nil {
		WriteError(w, http.StatusBadGateway, "تعذر رفع الصورة")
		return
	}
	_, _ = h.db.Exec(`UPDATE "PromoItem" SET "imageUrl" = $2 WHERE id = $1`, r.PathValue("id"), key)
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/customer-app/accounts — حسابات الزبائن ونشاطهم.
func (h *CustomerAppHandler) AdminAccounts(w http.ResponseWriter, r *http.Request) {
	rows := []struct {
		ID         string     `db:"id" json:"id"`
		Username   string     `db:"username" json:"username"`
		Name       string     `db:"name" json:"name"`
		Phone      string     `db:"phone" json:"phone"`
		Email      string     `db:"email" json:"email"`
		Verified   bool       `db:"verified" json:"verified"`
		Messages   int        `db:"messages" json:"messages"`
		Bookings   int        `db:"bookings" json:"bookings"`
		LastChatAt *time.Time `db:"lastChatAt" json:"lastChatAt"`
		CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
	}{}
	_ = h.db.Select(&rows, `SELECT a.id, a.username, a.name, a.phone, a.email, a."emailVerifiedAt" IS NOT NULL AS verified, a."createdAt",
		(SELECT count(*) FROM "CustomerChatMessage" m WHERE m."accountId" = a.id) AS messages,
		(SELECT max("createdAt") FROM "CustomerChatMessage" m WHERE m."accountId" = a.id) AS "lastChatAt",
		(SELECT count(*) FROM "Booking" b WHERE b."customerAccountId" = a.id) AS bookings
		FROM "CustomerAccount" a ORDER BY a."createdAt" DESC LIMIT 500`)
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/customer-app/accounts/{id}/chat
func (h *CustomerAppHandler) AdminChat(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.messages(r.PathValue("id"), 300))
}

// POST /api/customer-app/promos/suggest — ماتركس يقترح عرض (مسودة غير منشورة).
func (h *CustomerAppHandler) AdminSuggest(w http.ResponseWriter, r *http.Request) {
	if !h.matrix.Enabled() {
		WriteError(w, http.StatusServiceUnavailable, "ماتركس يحتاج مفتاح Anthropic")
		return
	}
	title, body, err := h.matrix.SuggestPromo()
	if err != nil {
		WriteError(w, http.StatusBadGateway, "ماتركس ما كدر يقترح هسه")
		return
	}
	id := uuid.NewString()
	_, _ = h.db.Exec(`INSERT INTO "PromoItem" (id, kind, title, body, active, source) VALUES ($1,'offer',$2,$3,false,'matrix')`, id, title, body)
	WriteJSON(w, http.StatusOK, map[string]string{"id": id, "title": title, "body": body})
}
