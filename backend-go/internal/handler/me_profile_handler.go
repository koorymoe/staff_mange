package handler

import (
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ 👤 ملفي الشخصي (دازاي 11-28) ═══
// «من يضغط الموظف على اسمه يدخل ع الملف الشخصي مالته ويشوف كلشي يخصه»: معلوماته (ويا راتبه هو
// بس)، دوامه وسجل حضوره، وخصوماته وغراماته ومنو خصمها. كلها عن نفسه — ماكو معرّف موظف بالطلب.

type MeProfileHandler struct {
	db         *sqlx.DB
	discipline *repository.DisciplineRecordRepository
}

func NewMeProfileHandler(db *sqlx.DB) *MeProfileHandler {
	return &MeProfileHandler{db: db, discipline: repository.NewDisciplineRecordRepository(db)}
}

type meInfo struct {
	Name            string   `db:"name" json:"name"`
	Username        string   `db:"username" json:"username"`
	Phone           *string  `db:"phone" json:"phone"`
	Email           *string  `db:"email" json:"email"`
	Role            string   `db:"role" json:"role"`
	JobTitle        *string  `db:"jobTitle" json:"jobTitle"`
	Certificate     *string  `db:"certificate" json:"certificate"`
	Salary          *float64 `db:"salary" json:"salary"`
	Shift           *string  `db:"shift" json:"shift"`
	ShiftStart      *string  `db:"shiftStart" json:"shiftStart"`
	ShiftEnd        *string  `db:"shiftEnd" json:"shiftEnd"`
	HireDate        *string  `db:"hireDate" json:"hireDate"`
	ContractEnd     *string  `db:"contractEndDate" json:"contractEnd"`
	ExperienceYears *float64 `db:"experienceYears" json:"experienceYears"`
	HasLicense      bool     `db:"hasDrivingLicense" json:"hasDrivingLicense"`
	IsLeader        bool     `db:"isLeader" json:"isLeader"`
	CreatedAt       string   `db:"createdAt" json:"createdAt"`
}

type meAttendance struct {
	Day      string     `db:"day" json:"day"`
	CheckIn  *time.Time `db:"checkIn" json:"checkIn"`
	CheckOut *time.Time `db:"checkOut" json:"checkOut"`
	Minutes  *int       `db:"minutes" json:"minutes"`
}

// GET /api/me/profile
func (h *MeProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var info meInfo
	if err := h.db.Get(&info, `SELECT name, username, phone, email, role::text AS role, "jobTitle", certificate, salary,
		shift::text AS shift, "shiftStart", "shiftEnd", "hireDate"::text AS "hireDate", "contractEndDate"::text AS "contractEndDate",
		"experienceYears", COALESCE("hasDrivingLicense", false) AS "hasDrivingLicense", COALESCE("isLeader", false) AS "isLeader",
		"createdAt"::date::text AS "createdAt"
		FROM "Employee" WHERE id = $1`, me); err != nil {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	att := []meAttendance{}
	_ = h.db.Select(&att, `SELECT baghdad_date(a."checkIn")::text AS day, a."checkIn", a."checkOut",
		CASE WHEN a."checkOut" IS NOT NULL THEN (EXTRACT(EPOCH FROM a."checkOut" - a."checkIn") / 60)::int END AS minutes
		FROM "Attendance" a WHERE a."employeeId" = $1 AND a."checkIn" IS NOT NULL AND a."checkIn" > now() - interval '45 days'
		ORDER BY a."checkIn" DESC`, me)
	ded, err := h.discipline.All(me)
	if err != nil {
		ded = []repository.DisciplineEntry{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"info": info, "attendance": att, "deductions": ded})
}

// PUT /api/me/email {email} — (11-28) الموظف يكتب إيميله بنفسه (فاضي = يشيله)
func (h *MeProfileHandler) SetEmail(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	e := strings.TrimSpace(in.Email)
	if e != "" {
		if a, err := mail.ParseAddress(e); err != nil || a.Address != e || len(e) > 120 {
			WriteError(w, http.StatusBadRequest, "الإيميل مو صحيح")
			return
		}
	}
	if _, err := h.db.Exec(`UPDATE "Employee" SET email = NULLIF($2, '') WHERE id = $1`, middleware.EmployeeIDFromContext(r), e); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"email": e})
}
