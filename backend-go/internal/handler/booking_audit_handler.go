package handler

import (
	"log"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type BookingAuditHandler struct {
	repo     *repository.BookingAuditRepository
	bookings *repository.BookingRepository
	notify   *repository.NotificationRepository
	invoices *repository.LeaderInvoiceRepository
	// إغلاق البلاغ بمخالفة انضباط — نعيد استعمال نظام الانضباط
	// الموجود بدل ما نبني عقوبة موازية.
	discipline *repository.DisciplineRepository
	employees  *repository.EmployeeRepository
	// permissions: حتى نعرف إذا المالك نطى المراقب مفتاح التدقيق
	// (`finance_audit`) — بلاها الرفض يبقى بالدور حصراً.
	permissions *repository.PermissionRepository
}

func NewBookingAuditHandler(r *repository.BookingAuditRepository, b *repository.BookingRepository, n *repository.NotificationRepository, inv *repository.LeaderInvoiceRepository, d *repository.DisciplineRepository, e *repository.EmployeeRepository) *BookingAuditHandler {
	return &BookingAuditHandler{repo: r, bookings: b, notify: n, invoices: inv, discipline: d, employees: e}
}

// SetPermissions يربط مستودع الصلاحيات بعد البناء — نفس نمط
// `SetNotifications` بباقي المعالجات، حتى ما نكسر كل نداءات البناء.
func (h *BookingAuditHandler) SetPermissions(p *repository.PermissionRepository) {
	h.permissions = p
}

// stampInvoiceVerdict ينزّل حكم التدقيق اليومي على فاتورة الحجز.
//
// طلب صاحب العمل: «الفواتير أطابقهن، من أطابقهن يروحون وين؟ فواتير
// بحاجة لاعتماد». قبل هذا، التدقيق اليومي جان يشتغل على الحجز بس
// والفاتورة تضل بلا حكم — فالمحاسب يدقق ٢٠ حجز وقائمة الفواتير ما
// تتحرك ولا خطوة، ويرجع يحكم عليهن وحدة وحدة من جديد.
//
// ⚠️ ما يعدّل فاتورة معتمدة (SetAuditVerdict ترفض) ولا يفشّل التدقيق
// إذا ماكو فاتورة أصلاً — أكو حجوزات تنتدقق بتقدير الإداري بلا فاتورة.
// hasAuditKey هل الموظف عنده مفتاح التدقيق الي ينطيه المالك بيده؟
//
// ⚠️ فشل القراءة **يضيّق ما يوسّع**: خطأ بقاعدة البيانات ما يصير
// يفتح قرار التدقيق لمن مو مخوّل (نفس مبدأ `canSeeAllBookings`).
func (h *BookingAuditHandler) hasAuditKey(employeeID string) bool {
	if h.permissions == nil || employeeID == "" {
		return false
	}
	has, err := h.permissions.HasPermission(employeeID, "finance_audit")
	return err == nil && has
}

// freeInvoiceForBooking يرجّع: هل أحدث فاتورة للحجز مؤشَّرة مجانية
// من الليدر، ومعرّفها.
//
// ⚠️ نفس اختيار «الأحدث» الي يستعمله `stampInvoiceVerdict` بالضبط —
// حتى ما نأشّر فاتورة ونختم ثانية.
func (h *BookingAuditHandler) freeInvoiceForBooking(bookingID string) (bool, string) {
	if h.invoices == nil {
		return false, ""
	}
	rows, err := h.invoices.ListByBooking(bookingID)
	if err != nil || len(rows) == 0 {
		return false, ""
	}
	latest := rows[len(rows)-1]
	return latest.IsFree, latest.ID
}

func (h *BookingAuditHandler) stampInvoiceVerdict(bookingID, verdict string, note *string, empID string, amount *float64) {
	if h.invoices == nil {
		return
	}
	rows, err := h.invoices.ListByBooking(bookingID)
	if err != nil || len(rows) == 0 {
		return
	}
	latest := rows[len(rows)-1] // ListByBooking مرتبة تصاعدياً بالإنشاء
	noteStr := ""
	if note != nil {
		noteStr = *note
	}
	if _, err := h.invoices.SetAuditVerdict(latest.ID, verdict, noteStr, empID, amount); err != nil {
		log.Printf("stamp invoice verdict (booking %s): %v", bookingID, err)
	}
}

// PUT /api/bookings/{id}/audit — قرار المحاسب.
//
// إما «مطابق» (ويأشر مدقق، ويشترط مبلغ)، أو بلاغ خطأ ينوجّه للمعني.
func (h *BookingAuditHandler) Audit(w http.ResponseWriter, r *http.Request) {
	var req model.AuditBookingRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	bookingID := r.PathValue("id")
	empID := middleware.EmployeeIDFromContext(r)

	// ⚠️⚠️ المراقب المدقق يشوف ولا يقرّر.
	//
	// المراقب ياخذ صلاحية «finance» افتراضياً بدوره
	// (RoleDefaultPermissions)، والحارس على هذا المسار يطلب نفس
	// الصلاحية بالضبط — فچان يقدر يضغط «مطابق» ويختم الفاتورة
	// ويرحّلها لطابور الاعتماد، يعني يسوي شغل المحاسب نفسه.
	// وهذا يكسر الفصل الي انبنى بفواتير الليدر: المراقب مفروض
	// **يراجع** قرار المحاسب مو يصدره.
	//
	// ⚠️ ونستعمل رفضاً صريحاً مو تسجيل مخالفة: الأزرار چانت
	// معروضة إله، فضغطه عليها مو محاولة تجاوز.
	// ⚠️⚠️ والاستثناء: **مفتاح المالك**. صاحب النظام طلب صراحةً
	// «أريد عنده خيار أيضاً يدقّق مثله مثل المحاسب» — فلمّا ينطي
	// المراقب صلاحية `finance_audit` بيده، الرفض هذا يصير **عائقاً
	// بلا معنى**: الشاشة تعرضله الأزرار (الصلاحية تفتحها) والخادم
	// يرجّعه ٤٠٣ — يعني منح ما ينفع بشي، وهو نفس العيب الي طلعنا منه.
	//
	// 🔴 والافتراضي يبقى الرفض: المراقب بلا مفتاح **يشوف ولا يقرّر**.
	if middleware.RoleFromContext(r) == "MONITOR" && !h.hasAuditKey(empID) {
		WriteError(w, http.StatusForbidden,
			"المراقب يراجع قرارات التدقيق ولا يصدرها — القرار للمحاسب، إلا إذا المالك نطاه صلاحية «تدقيق ومطابقة الحسابات»")
		return
	}

	switch req.Action {
	case model.AuditVerify:
		if err := h.repo.Verify(bookingID, req.AmountCollected, req.AdvancePaid); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		// «مطابق» يرحّل الفاتورة لطابور الاعتماد بحكمها مثبّت.
		h.stampInvoiceVerdict(bookingID, model.AuditVerdictMatched, req.Note, empID, req.AmountCollected)

	case model.AuditFree:
		// ═══ صيانة مجانية ═══
		//
		// الليدر يأشّر فاتورته مجانية بسبب من القائمة، وقتها صافيها
		// **صفر** — و«مطابق» يرفضها لأن حارس الأرباح يطلب مبلغاً
		// أكبر من صفر. فچان المحاسب إما يكتب مبلغاً ما انستلم، أو
		// يأشّر «غير مطابق» فتنفتح مخالفة على شغل سليم.
		//
		// 🔴 ولا بلاغ ولا إشعار رقابة/جودة هنا: المجانية **شغل
		// موثَّق بسببه**، مو خطأ.
		freeInvoice, invoiceID := h.freeInvoiceForBooking(bookingID)
		if err := h.repo.VerifyFree(bookingID, freeInvoice); err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		// ⚠️ لو المحاسب هو الي أشّرها (والليدر نساها) نأشّر الفاتورة
		// مجانية بعد — وإلا يصير الحجز مجانياً وفاتورته مو مجانية،
		// علمان متناقضان والمالك ما يعرف أي واحد يصدّق.
		if invoiceID != "" && !freeInvoice {
			reason := ""
			if req.FreeReasonID != nil {
				reason = strings.TrimSpace(*req.FreeReasonID)
			}
			if err := h.invoices.MarkFree(invoiceID, reason); err != nil {
				log.Printf("mark invoice free (booking %s): %v", bookingID, err)
			}
		}
		zero := 0.0
		h.stampInvoiceVerdict(bookingID, model.AuditVerdictFree, req.Note, empID, &zero)
		// 🔴 إشعار **للمالك وحده**: إسقاط مبلغ قرار مالي يستاهل
		// يُعرف منو أشّره ومتى — بلا ما ينحسب غلطاً على أحد.
		if h.notify != nil {
			if b, err := h.bookings.FindByID(bookingID); err == nil && b != nil {
				msg := "🎁 صيانة مجانية بالحجز " + b.Code + " — أشّرها المحاسب بالتدقيق"
				if req.Note != nil && strings.TrimSpace(*req.Note) != "" {
					msg += ": " + strings.TrimSpace(*req.Note)
				}
				_ = h.notify.CreateForRole("OWNER", "audit_issue", msg)
			}
		}

	case model.AuditMismatch, model.AuditPriceError:
		// المبلغ الي كتبه المحاسب ينحفظ حتى لو أشّر خطأ — هو الرقم
		// الصحيح من الفاتورة، والفرق يبين للي راح يتابع.
		if req.AmountCollected != nil || req.AdvancePaid != nil {
			_ = h.repo.SetAmount(bookingID, req.AmountCollected, req.AdvancePaid)
		}
		b, err := h.bookings.FindByID(bookingID)
		if err != nil || b == nil {
			WriteError(w, http.StatusNotFound, "الحجز غير موجود")
			return
		}
		issue, err := h.repo.RaiseIssue(bookingID, req.Action, req.Note, req.AmountCollected, b.AmountCollected, empID)
		if err != nil {
			log.Printf("raise audit issue: %v", err)
			WriteError(w, http.StatusBadRequest, "تعذر تسجيل البلاغ")
			return
		}
		// نفس الشي للأحكام السلبية: الفاتورة تحمل سبب رفضها، فالمراقب
		// يفتحها ويلگه ليش انتأشرت بلا ما يدور بالبلاغات.
		{
			v := model.AuditVerdictMismatch
			if req.Action == model.AuditPriceError {
				v = model.AuditVerdictPriceError
			}
			h.stampInvoiceVerdict(bookingID, v, req.Note, empID, req.AmountCollected)
		}
		if h.notify != nil {
			msg := "💸 " + issue.KindLabel + " بالحجز " + issue.BookingCode +
				" — سجّله " + issue.RaisedByName
			if issue.Note != nil && *issue.Note != "" {
				msg += ": " + *issue.Note
			}
			for _, role := range model.AuditRoutedRoles(req.Action) {
				_ = h.notify.CreateForRole(role, "audit_issue", msg)
			}
			_ = h.notify.CreateForRole("OWNER", "audit_issue", msg)
		}
		WriteJSON(w, http.StatusCreated, issue)
		return

	default:
		WriteError(w, http.StatusBadRequest, "نوع القرار غير معروف")
		return
	}

	updated, err := h.bookings.FindByID(bookingID)
	if err != nil || updated == nil {
		WriteError(w, http.StatusNotFound, "الحجز غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, updated)
}

// GET /api/audit-issues?status=OPEN — بلاغات الأخطاء.
//
// كل واحد يشوف الي يخصه: الجودة تشوف «غير مطابق»، والإداري يشوف
// «خطأ سعر»، والرقابة والمالك يشوفون الاثنين.
func (h *BookingAuditHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	var kinds []string
	switch role {
	case "QUALITY_ENGINEER":
		kinds = []string{model.AuditMismatch}
	case "HR_COORDINATOR":
		kinds = []string{model.AuditPriceError}
	}
	// ═══ المحاسب مو مراقب ═══
	// «أخطاء الفواتير تظهر إله كمحاسب — هو أرسلهن. بس تظهر للمراقب
	// كتدقيق حتى يتأكد من الليدر ليش عنده أخطاء.»
	// المحاسب يشوف **صادره** هو بس؛ المراقب والمدير يشوفون الكل مع
	// اسم الليدر. بدون هذا الاثنين جانوا يشوفون نفس القائمة بالضبط،
	// فالمحاسب يحسبها شغل عليه وهو أصلاً الي سجّلها.
	raisedBy := ""
	if role == "FINANCE" {
		raisedBy = middleware.EmployeeIDFromContext(r)
	}
	rows, err := h.repo.List(strings.TrimSpace(r.URL.Query().Get("status")), kinds, raisedBy)
	if err != nil {
		log.Printf("list audit issues: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البلاغات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/audit-issues/{id}/resolve
func (h *BookingAuditHandler) ResolveIssue(w http.ResponseWriter, r *http.Request) {
	// البلاغ ينغلق من الجهة الي انوجّهله بس. بدون هالفحص أي موظف مسجّل
	// دخول يقدر يغلق البلاغ — يعني نفس الي سبّب خطأ السعر يغلق البلاغ
	// المرفوع ضده، وكل فايدة التوجيه تروح.
	kind, err := h.repo.KindOf(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "البلاغ غير موجود")
		return
	}
	role := middleware.RoleFromContext(r)
	allowed := role == "ADMIN" || role == "OWNER"
	for _, rt := range model.AuditRoutedRoles(kind) {
		if rt == role {
			allowed = true
			break
		}
	}
	if !allowed {
		WriteError(w, http.StatusForbidden, "هذا البلاغ ما ينغلق إلا من "+model.AuditRoutedLabel(kind))
		return
	}
	// ⚠️⚠️ ما ينغلق إلا بإجراء. قبلها چان سطراً واحداً يأشّره
	// «محلول» — بلا منو ولا ليش ولا أثر على أحد، فالبلاغ يختفي
	// وينتهي. صاحب النظام سأل «وين يروح؟» والجواب چان: ماكو مكان.
	var req model.ResolveAuditIssueRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if _, ok := model.AuditActionLabels[req.Action]; !ok {
		WriteError(w, http.StatusBadRequest,
			"لازم تختار إجراء: مخالفة انضباط، أو تأكيد إنه ماكو خطأ")
		return
	}
	// ⚠️ السبب إجباري بالحالتين: بدونه «تأكدت ماكو خطأ» تصير باباً
	// خلفياً للإغلاق الروتيني — يضغط ويسكّر بلا ما يقرا.
	if strings.TrimSpace(req.Reason) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الإغلاق")
		return
	}

	issue, err := h.repo.Find(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "البلاغ غير موجود")
		return
	}

	byID := middleware.EmployeeIDFromContext(r)
	byName := ""
	if e, err := h.employees.FindByID(byID); err == nil && e != nil {
		byName = e.Name
	}

	// مخالفة الانضباط تنسجّل **قبل** الإغلاق: لو فشلت ما نريد بلاغاً
	// مسكّراً بلا العقوبة الي انسكّر على أساسها.
	if req.Action == model.AuditActionPenalize {
		leaderID, err := h.repo.LeaderIDForBooking(issue.BookingID)
		if err != nil || leaderID == "" {
			WriteError(w, http.StatusBadRequest, "ما لكيت ليدر مرتبط بهذا الحجز")
			return
		}
		points := req.Points
		if points <= 0 {
			points = 1
		}
		reason := model.AuditIssueLabels[issue.Kind] + " — " + req.Reason
		bookingID := issue.BookingID
		if _, _, _, err := h.discipline.Penalize(leaderID, "AUDIT_ISSUE", reason, &bookingID, points); err != nil {
			WriteError(w, http.StatusBadRequest, "تعذر تسجيل المخالفة: "+err.Error())
			return
		}
	}

	if err := h.repo.Resolve(r.PathValue("id"), byID, byName, req.Action, req.Reason); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إغلاق البلاغ")
		return
	}

	// ⚠️ إشعار للمحاسب الي سجّل البلاغ: بدونه يسجّل بلاغاً وما يدري
	// أبداً شنو صار بيه.
	if h.notify != nil {
		msg := "بلاغ " + issue.BookingCode + " (" + model.AuditIssueLabels[issue.Kind] +
			") انغلق — " + model.AuditActionLabels[req.Action] + ": " + req.Reason +
			" · أغلقه " + byName
		_ = h.notify.Create(issue.RaisedByID, "audit_issue_closed", msg)
		if leaderID, err := h.repo.LeaderIDForBooking(issue.BookingID); err == nil && leaderID != "" {
			_ = h.notify.Create(leaderID, "audit_issue_closed", msg)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
