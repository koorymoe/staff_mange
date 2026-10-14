package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// ═══ تطبيق الزبائن — الكتالوك وتفاصيل الحجز (تصميم (ع) 10-14) ═══
//
// بطاقة كل خدمة (صورة، وصف، مميزات، «يبدأ من») يكتبها المالك. التقييم
// حقيقي: معدّل تقييمات الزبائن لحجوزات الخدمة — والفارغ ما يطلع.

type appFeature struct {
	Icon string `json:"icon"`
	Text string `json:"text"`
}

type serviceCard struct {
	ID          string          `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Category    *string         `db:"category" json:"category"`
	Description *string         `db:"description" json:"description"`
	FeaturesRaw json.RawMessage `db:"features" json:"features"`
	PriceFrom   *float64        `db:"priceFrom" json:"priceFrom"`
	HasImage    bool            `db:"hasImage" json:"hasImage"`
	Color       *string         `db:"color" json:"color"`
	Visible     bool            `db:"visible" json:"visible"`
	Ord         int             `db:"ord" json:"ord"`
	Rating      *float64        `db:"rating" json:"rating"`
	RatingCount int             `db:"ratingCount" json:"ratingCount"`
	Favorite    bool            `db:"favorite" json:"favorite"`
}

// الخدمات الإدارية الداخلية ما تطلع للزبون
const appServiceFilter = `s.category IS DISTINCT FROM 'إدارية'`

func (h *CustomerAppHandler) serviceCards(accountID, where string, args ...any) ([]serviceCard, error) {
	rows := []serviceCard{}
	q := `SELECT s.id, s.name, s.category, sa.description, COALESCE(sa.features, '[]'::jsonb) AS features, sa."priceFrom",
			COALESCE(sa."imageUrl", '') <> '' AS "hasImage", sa.color, COALESCE(sa.visible, true) AS visible, COALESCE(sa.ord, 100) AS ord,
			r.rating, COALESCE(r.n, 0) AS "ratingCount",
			EXISTS (SELECT 1 FROM "CustomerFavorite" f WHERE f."serviceId" = s.id AND f."accountId" = $1) AS favorite
		FROM "Service" s
		LEFT JOIN "ServiceApp" sa ON sa."serviceId" = s.id
		LEFT JOIN LATERAL (
			SELECT round(avg(cr.stars)::numeric, 1)::float8 AS rating, count(*) AS n FROM "CustomerRating" cr
			JOIN "Booking" b ON b.id = cr."bookingId"
			WHERE b."serviceId" = s.id OR EXISTS (SELECT 1 FROM "BookingService" bs WHERE bs."bookingId" = b.id AND bs."serviceId" = s.id)
		) r ON true
		WHERE ` + appServiceFilter + ` AND ` + where + `
		ORDER BY COALESCE(sa.ord, 100), s.name`
	err := h.db.Select(&rows, q, append([]any{accountID}, args...)...)
	return rows, err
}

// GET /api/c/services
func (h *CustomerAppHandler) Services(w http.ResponseWriter, r *http.Request) {
	rows, err := h.serviceCards(current(r).ID, `COALESCE(sa.visible, true)`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الخدمات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/c/services/{id}
func (h *CustomerAppHandler) ServiceDetail(w http.ResponseWriter, r *http.Request) {
	rows, err := h.serviceCards(current(r).ID, `s.id = $2`, r.PathValue("id"))
	if err != nil || len(rows) == 0 {
		WriteError(w, http.StatusNotFound, "الخدمة غير موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, rows[0])
}

// GET /api/c/img/s/{id} — صورة الخدمة (محتوى ترويجي عام)
func (h *CustomerAppHandler) ServiceImage(w http.ResponseWriter, r *http.Request) {
	var key string
	if err := h.db.Get(&key, `SELECT COALESCE("imageUrl", '') FROM "ServiceApp" WHERE "serviceId" = $1`, r.PathValue("id")); err != nil || key == "" {
		http.NotFound(w, r)
		return
	}
	h.serveKey(w, r, key, "public, max-age=86400")
}

func (h *CustomerAppHandler) serveKey(w http.ResponseWriter, r *http.Request, key, cache string) {
	data, ct, err := h.store.Get(r.Context(), key)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", cache)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

// ─── المفضلة ───

func (h *CustomerAppHandler) Favorites(w http.ResponseWriter, r *http.Request) {
	rows, err := h.serviceCards(current(r).ID, `EXISTS (SELECT 1 FROM "CustomerFavorite" f WHERE f."serviceId" = s.id AND f."accountId" = $1)`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المفضلة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *CustomerAppHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`INSERT INTO "CustomerFavorite" ("accountId", "serviceId") VALUES ($1, $2) ON CONFLICT DO NOTHING`, current(r).ID, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *CustomerAppHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "CustomerFavorite" WHERE "accountId" = $1 AND "serviceId" = $2`, current(r).ID, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── طلب خدمة من النموذج ───

// POST /api/c/bookings (multipart): serviceId, city, date, time, address, notes, photos[]
// يتسجّل باسم ماتركس — هو الي يستلم طلبات التطبيق (قرار (ع)).
func (h *CustomerAppHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	a := current(r)
	if err := r.ParseMultipartForm(4 * storage.MaxFileBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر قراءة الطلب — يمكن الصور كبيرة")
		return
	}
	v := func(k string) string { return strings.TrimSpace(r.FormValue(k)) }
	svc, city, addr := v("serviceId"), v("city"), v("address")
	if svc == "" || city == "" || addr == "" {
		WriteError(w, http.StatusBadRequest, "اختار الخدمة والمحافظة واكتب العنوان")
		return
	}
	var ok bool
	_ = h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Service" s LEFT JOIN "ServiceApp" sa ON sa."serviceId" = s.id
		WHERE s.id = $1 AND `+appServiceFilter+` AND COALESCE(sa.visible, true))`, svc)
	if !ok {
		WriteError(w, http.StatusBadRequest, "الخدمة غير متاحة")
		return
	}
	files := r.MultipartForm.File["photos"]
	if len(files) > 4 {
		WriteError(w, http.StatusBadRequest, "أقصى شي 4 صور")
		return
	}
	// نقرا الصور ونتأكد منها قبل ما نسجّل الحجز
	type photo struct {
		data []byte
		ct   string
	}
	photos := []photo{}
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			continue
		}
		data, err := storage.ReadLimited(f, storage.MaxFileBytes)
		f.Close()
		if err != nil {
			WriteError(w, http.StatusRequestEntityTooLarge, "صورة كبيرة — أقصى حجم 5MB")
			return
		}
		ct := storage.SniffContentType(data)
		if !strings.HasPrefix(ct, "image/") || !storage.AllowedContentTypes[ct] {
			WriteError(w, http.StatusBadRequest, "الصور لازم JPG أو PNG أو WEBP")
			return
		}
		photos = append(photos, photo{data, ct})
	}
	when := strings.TrimSpace(v("date") + " " + v("time"))
	if a.City == nil || *a.City == "" {
		_, _ = h.db.Exec(`UPDATE "CustomerAccount" SET city = $2 WHERE id = $1`, a.ID, city)
	}
	notes := v("notes")
	if len(photos) > 0 {
		notes = strings.TrimSpace(notes + "\n📷 أرفق " + strconv.Itoa(len(photos)) + " صورة")
	}
	code, err := h.createBookingFromMatrix(a.ID, service.CustomerBookingDraft{
		ServiceIDs: []string{svc}, Address: city + " — " + addr, PreferredTime: when, Notes: notes,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	var id string
	_ = h.db.Get(&id, `SELECT id FROM "Booking" WHERE code = $1`, code)
	for _, p := range photos {
		key := storage.NewKey("customer-app", p.ct)
		if err := h.store.Put(r.Context(), key, p.data, p.ct); err == nil {
			_, _ = h.db.Exec(`INSERT INTO "CustomerBookingPhoto" (id, "bookingId", key) VALUES ($1,$2,$3)`, uuid.NewString(), id, key)
		}
	}
	WriteJSON(w, http.StatusOK, map[string]string{"id": id, "code": code})
}

// ─── تفاصيل الحجز ───

func (h *CustomerAppHandler) owns(accountID, bookingID string) bool {
	var ok bool
	_ = h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Booking" b WHERE b.id = $2 AND (b."customerAccountId" = $1
		OR b."customerId" = (SELECT "customerId" FROM "CustomerAccount" WHERE id = $1)))`, accountID, bookingID)
	return ok
}

type appTech struct {
	ID          string   `db:"id" json:"-"`
	Name        string   `db:"name" json:"name"`
	HasPhoto    bool     `db:"hasPhoto" json:"hasPhoto"`
	Rating      *float64 `db:"rating" json:"rating"`
	RatingCount int      `db:"ratingCount" json:"ratingCount"`
}

func (h *CustomerAppHandler) techOf(bookingID string) *appTech {
	var t appTech
	err := h.db.Get(&t, `SELECT e.id, e.name, COALESCE(e."photoUrl", '') <> '' AS "hasPhoto",
			(SELECT round(avg(cr.stars)::numeric, 1)::float8 FROM "CustomerRating" cr JOIN "BookingAssignment" x ON x."bookingId" = cr."bookingId" WHERE x."employeeId" = e.id) AS rating,
			(SELECT count(*) FROM "CustomerRating" cr JOIN "BookingAssignment" x ON x."bookingId" = cr."bookingId" WHERE x."employeeId" = e.id) AS "ratingCount"
		FROM "BookingAssignment" ba JOIN "Employee" e ON e.id = ba."employeeId"
		WHERE ba."bookingId" = $1 ORDER BY ba.role, ba."createdAt" LIMIT 1`, bookingID)
	if err != nil {
		return nil
	}
	return &t
}

// GET /api/c/bookings/{id}
func (h *CustomerAppHandler) BookingDetail(w http.ResponseWriter, r *http.Request) {
	a, id := current(r), r.PathValue("id")
	if !h.owns(a.ID, id) {
		WriteError(w, http.StatusForbidden, "هذا الحجز مو إلك")
		return
	}
	var b struct {
		ID              string     `db:"id" json:"id"`
		Code            string     `db:"code" json:"code"`
		Status          string     `db:"status" json:"status"`
		ServiceID       *string    `db:"serviceId" json:"serviceId"`
		Address         *string    `db:"address" json:"address"`
		ScheduledAt     *time.Time `db:"scheduledAt" json:"scheduledAt"`
		ConfirmedAt     *time.Time `db:"confirmedAt" json:"confirmedAt"`
		StartedAt       *time.Time `db:"startedAt" json:"startedAt"`
		CompletedAt     *time.Time `db:"completedAt" json:"completedAt"`
		CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
		QuotedPrice     *float64   `db:"quotedPrice" json:"quotedPrice"`
		AmountCollected *float64   `db:"amountCollected" json:"amountCollected"`
		FromApp         bool       `db:"fromApp" json:"fromApp"`
	}
	if err := h.db.Get(&b, `SELECT id, code, status, "serviceId", address, "scheduledAt", "confirmedAt", "startedAt", "completedAt", "createdAt",
			"quotedPrice"::float8 AS "quotedPrice", "amountCollected"::float8 AS "amountCollected", (source = 'matrix_app') IS TRUE AS "fromApp"
		FROM "Booking" WHERE id = $1`, id); err != nil {
		WriteError(w, http.StatusNotFound, "الحجز غير موجود")
		return
	}
	var services string
	_ = h.db.Get(&services, `SELECT COALESCE((SELECT string_agg(s.name, '، ') FROM "BookingService" bs JOIN "Service" s ON s.id = bs."serviceId" WHERE bs."bookingId" = $1),
		(SELECT s.name FROM "Service" s JOIN "Booking" b ON b."serviceId" = s.id WHERE b.id = $1), '')`, id)
	var description *string
	if b.ServiceID != nil {
		_ = h.db.Get(&description, `SELECT description FROM "ServiceApp" WHERE "serviceId" = $1`, *b.ServiceID)
	}
	var mine *struct {
		Stars   int     `db:"stars" json:"stars"`
		Comment *string `db:"comment" json:"comment"`
	}
	var rt struct {
		Stars   int     `db:"stars" json:"stars"`
		Comment *string `db:"comment" json:"comment"`
	}
	if err := h.db.Get(&rt, `SELECT stars, comment FROM "CustomerRating" WHERE "bookingId" = $1`, id); err == nil {
		mine = &rt
	}
	label := map[string]string{"PENDING": "بانتظار التثبيت", "CONFIRMED": "مجدول", "WAITING": "بالانتظار", "IN_PROGRESS": "قيد التنفيذ",
		"PARTIAL": "منجز جزئياً", "COMPLETED": "مكتمل", "CANCELLED": "ملغي"}[b.Status]
	WriteJSON(w, http.StatusOK, map[string]any{"booking": b, "statusLabel": label, "services": services, "description": description,
		"tech": h.techOf(id), "myRating": mine})
}

// GET /api/c/bookings/{id}/tech-photo — صورة الفني، لصاحب الحجز بس
func (h *CustomerAppHandler) TechPhoto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.owns(current(r).ID, id) {
		http.NotFound(w, r)
		return
	}
	t := h.techOf(id)
	if t == nil {
		http.NotFound(w, r)
		return
	}
	var url string
	_ = h.db.Get(&url, `SELECT COALESCE("photoUrl", '') FROM "Employee" WHERE id = $1`, t.ID)
	key := url
	if i := strings.Index(key, "/api/files/"); i >= 0 {
		key = key[i+len("/api/files/"):]
	}
	if i := strings.Index(key, "?"); i >= 0 {
		key = key[:i]
	}
	if key == "" || strings.HasPrefix(key, "http") {
		http.NotFound(w, r)
		return
	}
	h.serveKey(w, r, key, "private, max-age=3600")
}

// POST /api/c/bookings/{id}/rate {stars, comment}
func (h *CustomerAppHandler) Rate(w http.ResponseWriter, r *http.Request) {
	a, id := current(r), r.PathValue("id")
	if !h.owns(a.ID, id) {
		WriteError(w, http.StatusForbidden, "هذا الحجز مو إلك")
		return
	}
	var b struct {
		Stars   int    `json:"stars"`
		Comment string `json:"comment"`
	}
	_ = DecodeJSON(r, &b)
	if b.Stars < 1 || b.Stars > 5 {
		WriteError(w, http.StatusBadRequest, "اختار من 1 إلى 5 نجوم")
		return
	}
	var status string
	_ = h.db.Get(&status, `SELECT status FROM "Booking" WHERE id = $1`, id)
	if status != "COMPLETED" {
		WriteError(w, http.StatusBadRequest, "التقييم بعد ما تخلص الخدمة")
		return
	}
	var comment *string
	if c := strings.TrimSpace(b.Comment); c != "" {
		if len([]rune(c)) > 500 {
			c = string([]rune(c)[:500])
		}
		comment = &c
	}
	if _, err := h.db.Exec(`INSERT INTO "CustomerRating" (id, "bookingId", "accountId", stars, comment) VALUES ($1,$2,$3,$4,$5)`,
		uuid.NewString(), id, a.ID, b.Stars, comment); err != nil {
		WriteError(w, http.StatusConflict, "قيّمت هذا الحجز من قبل")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ─── الفروع ───

type appBranch struct {
	ID      string  `db:"id" json:"id"`
	Name    string  `db:"name" json:"name"`
	Address *string `db:"address" json:"address"`
	Phone   *string `db:"phone" json:"phone"`
	MapURL  *string `db:"mapUrl" json:"mapUrl"`
	Ord     int     `db:"ord" json:"ord"`
}

func (h *CustomerAppHandler) Branches(w http.ResponseWriter, r *http.Request) {
	rows := []appBranch{}
	_ = h.db.Select(&rows, `SELECT id, name, address, phone, "mapUrl", ord FROM "AppBranch" ORDER BY ord, name`)
	WriteJSON(w, http.StatusOK, rows)
}

// ═══ إدارة المالك ═══

// GET /api/customer-app/services — كل الخدمات ببطاقاتها (حتى المخفية)
func (h *CustomerAppHandler) AdminServices(w http.ResponseWriter, r *http.Request) {
	rows, err := h.serviceCards("", `true`)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/customer-app/services/{id}
func (h *CustomerAppHandler) AdminSaveService(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Description *string      `json:"description"`
		Features    []appFeature `json:"features"`
		PriceFrom   *float64     `json:"priceFrom"`
		Visible     bool         `json:"visible"`
		Ord         int          `json:"ord"`
		Color       *string      `json:"color"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	clean := []appFeature{}
	for _, f := range b.Features {
		if t := strings.TrimSpace(f.Text); t != "" && len(clean) < 6 {
			clean = append(clean, appFeature{Icon: strings.TrimSpace(f.Icon), Text: t})
		}
	}
	fj, _ := json.Marshal(clean)
	if b.PriceFrom != nil && *b.PriceFrom <= 0 {
		b.PriceFrom = nil
	}
	_, err := h.db.Exec(`INSERT INTO "ServiceApp" ("serviceId", description, features, "priceFrom", visible, ord, color, "updatedAt")
		VALUES ($1,$2,$3,$4,$5,$6,$7, now())
		ON CONFLICT ("serviceId") DO UPDATE SET description=$2, features=$3, "priceFrom"=$4, visible=$5, ord=$6, color=$7, "updatedAt"=now()`,
		r.PathValue("id"), b.Description, string(fj), b.PriceFrom, b.Visible, b.Ord, b.Color)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/customer-app/services/{id}/image
func (h *CustomerAppHandler) AdminServiceImage(w http.ResponseWriter, r *http.Request) {
	key, err := h.readImage(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, _ = h.db.Exec(`INSERT INTO "ServiceApp" ("serviceId", "imageUrl") VALUES ($1,$2)
		ON CONFLICT ("serviceId") DO UPDATE SET "imageUrl" = $2, "updatedAt" = now()`, r.PathValue("id"), key)
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// readImage يقرا صورة multipart (file) ويرفعها — يرجّع مفتاحها.
func (h *CustomerAppHandler) readImage(r *http.Request) (string, error) {
	if err := r.ParseMultipartForm(storage.MaxFileBytes); err != nil {
		return "", errors.New("تعذر قراءة الصورة")
	}
	f, _, err := r.FormFile("file")
	if err != nil {
		return "", errors.New("ماكو صورة")
	}
	defer f.Close()
	data, err := storage.ReadLimited(f, storage.MaxFileBytes)
	if err != nil {
		return "", err
	}
	ct := storage.SniffContentType(data)
	if !strings.HasPrefix(ct, "image/") || !storage.AllowedContentTypes[ct] {
		return "", errors.New("لازم صورة JPG/PNG/WEBP")
	}
	key := storage.NewKey("customer-app", ct)
	if err := h.store.Put(r.Context(), key, data, ct); err != nil {
		return "", errors.New("تعذر رفع الصورة")
	}
	return key, nil
}

// الفروع — المالك
func (h *CustomerAppHandler) AdminSaveBranch(w http.ResponseWriter, r *http.Request) {
	var b appBranch
	if err := DecodeJSON(r, &b); err != nil || strings.TrimSpace(b.Name) == "" {
		WriteError(w, http.StatusBadRequest, "اسم الفرع مطلوب")
		return
	}
	id := r.PathValue("id")
	var err error
	if id == "" {
		_, err = h.db.Exec(`INSERT INTO "AppBranch" (id, name, address, phone, "mapUrl", ord) VALUES ($1,$2,$3,$4,$5,$6)`,
			uuid.NewString(), strings.TrimSpace(b.Name), b.Address, b.Phone, b.MapURL, b.Ord)
	} else {
		_, err = h.db.Exec(`UPDATE "AppBranch" SET name=$2, address=$3, phone=$4, "mapUrl"=$5, ord=$6 WHERE id=$1`,
			id, strings.TrimSpace(b.Name), b.Address, b.Phone, b.MapURL, b.Ord)
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *CustomerAppHandler) AdminDeleteBranch(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "AppBranch" WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/customer-app/ratings — تقييمات الزبائن
func (h *CustomerAppHandler) AdminRatings(w http.ResponseWriter, r *http.Request) {
	rows := []struct {
		Code      string     `db:"code" json:"code"`
		Customer  string     `db:"customer" json:"customer"`
		Service   string     `db:"service" json:"service"`
		Tech      *string    `db:"tech" json:"tech"`
		Stars     int        `db:"stars" json:"stars"`
		Comment   *string    `db:"comment" json:"comment"`
		CreatedAt time.Time  `db:"createdAt" json:"createdAt"`
		Done      *time.Time `db:"completedAt" json:"completedAt"`
	}{}
	err := h.db.Select(&rows, `SELECT b.code, a.name AS customer, COALESCE(s.name, '') AS service,
			(SELECT e.name FROM "BookingAssignment" ba JOIN "Employee" e ON e.id = ba."employeeId" WHERE ba."bookingId" = b.id ORDER BY ba.role LIMIT 1) AS tech,
			cr.stars, cr.comment, cr."createdAt", b."completedAt"
		FROM "CustomerRating" cr JOIN "Booking" b ON b.id = cr."bookingId" JOIN "CustomerAccount" a ON a.id = cr."accountId"
		LEFT JOIN "Service" s ON s.id = b."serviceId" ORDER BY cr."createdAt" DESC LIMIT 300`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
