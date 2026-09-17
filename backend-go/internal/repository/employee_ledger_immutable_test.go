package repository

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ═══ حارس: دفتر ذمة الموظف ما ينعدّل ولا ينمسح ═══
//
// الدفتر يحمل **أرقام فلوس** تنبني عليها أحكام ماتركس والخصم من الراتب.
// صف ينعدّل بهدوء يعني الدليل ينقلب لتزوير ومحد يعرف.
//
// الحماية طبقتان، وهذا الملف يحرس الاثنتين:
//
//	١. **قاعدة البيانات** — مُشغّلان يرفضان UPDATE و DELETE. هاي
//	   الحماية الحقيقية لأنها تمسك حتى psql بالإيد والسكربتات.
//	٢. **الكود** — ماكو أي مسار يحاول يعدّل أصلاً.
//
// ⚠️ والاختبار البنيوي هو الأهم: يفشل **يوم ينكتب مسار التعديل** مو
// يوم يكتشف المالك إن رقماً انتغيّر.

// ⚠️ الاستثناء الوحيد: ملف الترحيل نفسه يذكر UPDATE/DELETE داخل نص
// المُشغّل الي **يمنعهن**. أي ملف ثاني يذكرهن = علّة.
var ledgerMutationExempt = map[string]string{
	"schema_employee_ledger.go":         "الترحيل نفسه — يعرّف المُشغّلات الي تمنع التعديل",
	"employee_ledger_immutable_test.go": "هذا الحارس",
	"employee_ledger_repository.go":     "يترجم رسالة رفض القاعدة للعربية",
}

// TestNoCodePathMutatesLedger يقرا الباك إند كله ويتأكد إن ولا ملف
// يحاول يعدّل أو يمسح قيداً بالدفتر.
func TestNoCodePathMutatesLedger(t *testing.T) {
	// UPDATE "EmployeeLedgerEntry" ... أو DELETE FROM "EmployeeLedgerEntry"
	mutate := regexp.MustCompile(`(?i)(UPDATE\s+"EmployeeLedgerEntry"|DELETE\s+FROM\s+"EmployeeLedgerEntry")`)

	root := ".."
	found := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if _, ok := ledgerMutationExempt[filepath.Base(path)]; ok {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if mutate.Match(src) {
			found = append(found, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("مشي الملفات: %v", err)
	}
	if len(found) > 0 {
		t.Fatalf("أكو مسار يعدّل دفتر الذمة — والدفتر دليل مالي ما ينعدّل:\n  %s\n"+
			"التصحيح يصير بقيد عكسي (Reverse) مو بتعديل.", strings.Join(found, "\n  "))
	}
}

// TestMigrationDeclaresImmutability يتأكد إن الترحيل فعلاً يحط
// المُشغّلات والقيود — مو بس يعرّف الجدول.
//
// ⚠️ بدون هذا، أحد يگدر يشيل المُشغّل بترحيل لاحق وكل شي يبقى يبني
// ويمر، والحماية تختفي بهدوء.
func TestMigrationDeclaresImmutability(t *testing.T) {
	src, err := os.ReadFile("../database/schema_employee_ledger.go")
	if err != nil {
		t.Fatalf("قراءة الترحيل: %v", err)
	}
	s := string(src)

	required := map[string]string{
		"employee_ledger_no_update": "مُشغّل يمنع التعديل",
		"employee_ledger_no_delete": "مُشغّل يمنع الحذف",
		// 🔴 انكشفت بالتجربة: مُشغّلا الصف ما يمسكان TRUNCATE، وهي
		// مسحت الدفتر كله بأمر واحد. شيله = انهيار الحماية كلها.
		"employee_ledger_no_truncate":             "مُشغّل يمنع TRUNCATE — الصف وحده ما يمسكها",
		"FOR EACH STATEMENT":                      "منع TRUNCATE لازم يكون على مستوى الجملة مو الصف",
		"employee_ledger_collected_needs_booking": "التحصيل لازم ينربط بحجز — بدونه ترجع الفلوس اليتيمة",
		"employee_ledger_reversal_needs_note":     "التصحيح لازم إله سبب",
		"EmployeeLedgerEntry_reverse_unique":      "القيد ما ينعكس مرتين",
		"amount > 0":                              "المبلغ موجب — الاتجاه من kind مو من الإشارة",
	}
	for needle, why := range required {
		if !strings.Contains(s, needle) {
			t.Errorf("الترحيل ما بيه %q — %s", needle, why)
		}
	}
}

// TestLedgerErrorsAreArabic يتأكد إن رفض القاعدة يوصل للمستخدم بعربية
// مفهومة مو بنص Postgres.
func TestLedgerErrorsAreArabic(t *testing.T) {
	cases := map[string]string{
		"employee_ledger_collected_needs_booking": "ينربط بحجز",
		"employee_ledger_reversal_needs_note":     "سبب مكتوب",
		"EmployeeLedgerEntry_amount_check":        "أكبر من صفر",
		"EmployeeLedgerEntry_kind_check":          "مو معروف",
		"EmployeeLedgerEntry_reverse_unique":      "ما ينصحّح مرتين",
		"ما ينعدّل ولا ينمسح":                     "قيد عكسي",
	}
	for raw, want := range cases {
		got := translateLedgerError(&stubErr{raw}).Error()
		if !strings.Contains(got, want) {
			t.Errorf("رسالة %q طلعت %q — المتوقع تحتوي %q", raw, got, want)
		}
	}
	// خطأ مو معروف يمر كما هو بدل ما ينخفي وراء رسالة عامة.
	other := &stubErr{"connection refused"}
	if translateLedgerError(other).Error() != "connection refused" {
		t.Error("الخطأ غير المعروف لازم يمر كما هو حتى ما ينضيع بالتشخيص")
	}
}

type stubErr struct{ s string }

func (e *stubErr) Error() string { return e.s }

// ═══ اختبار القاعدة الحية ═══
//
// يشتغل بس إذا DATABASE_URL موجود — نفس نمط بقية حرّاس هذا المجلد.
// يثبت إن المُشغّلات شغّالة فعلاً مو بس مكتوبة بالترحيل.
func TestLedgerImmutableOnLiveDB(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	// ⚠️ **الثلاثة لازم يكونون موجودين**. مُشغّلا الصف ما يمسكان
	// TRUNCATE — انكشفت بالتجربة: TRUNCATE مسحت الدفتر كله وهي تمر
	// من فوقهم. فحص وجودهم بالاسم يمسك أي ترحيل لاحق يشيل واحداً.
	for _, tg := range []string{
		"employee_ledger_no_update",
		"employee_ledger_no_delete",
		"employee_ledger_no_truncate",
	} {
		var exists bool
		if err := db.Get(&exists,
			`SELECT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = $1)`, tg); err != nil {
			t.Fatalf("فحص المُشغّل %s: %v", tg, err)
		}
		if !exists {
			t.Errorf("مُشغّل %s مو موجود بالقاعدة — الدفتر مكشوف", tg)
		}
	}

	// ═══ محاولات فعلية داخل معاملة تتراجع ═══
	//
	// ⚠️ لازم تلمس **صفاً حقيقياً**: `WHERE false` ما تلمس ولا صف
	// فالمُشغّل ما يشتغل أصلاً، والاختبار يمر وهو ما فحص شي. هاي
	// الغلطة صارت بأول نسخة من الاختبار.
	tx, err := db.Beginx()
	if err != nil {
		t.Fatalf("فتح معاملة: %v", err)
	}
	defer tx.Rollback() //nolint:errcheck // معاملة فحص — ما ننوي نحفظها

	var probeID string
	err = tx.Get(&probeID, `SELECT id FROM "EmployeeLedgerEntry" LIMIT 1`)
	if err != nil {
		t.Skip("الدفتر فاضي — ماكو صف نجرّب عليه")
	}

	if _, err := tx.Exec(
		`UPDATE "EmployeeLedgerEntry" SET amount = amount + 1 WHERE id = $1`, probeID,
	); err == nil {
		t.Error("التعديل على صف حقيقي انقبل — مُشغّل المنع مو شغّال")
	} else if !strings.Contains(err.Error(), "ما ينعدّل") {
		t.Errorf("التعديل انرفض بسبب ثاني مو المُشغّل: %v", err)
	}
}
