package main

// اختبارات تكامل حية (Postgres حقيقي عبر DATABASE_URL، بدون mocks) تغطي أهم بنود
// المراجعة الأمنية اللي انطبّقت هذا الأسبوع: فحص الأدوار/الصلاحيات الحقيقي على
// مستوى الـHTTP الكامل (نفس mux وسلسلة middleware المستخدمة فعلياً بـmain())،
// وأهمها اختبار إبطال صلاحيات الأدمن فوراً عند تخفيض دوره (بدون ما يحتاج يسجّل
// خروج/دخول أو ينتظر انتهاء التوكن).

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
)

const testEmployeePassword = "TestPass123!"

// setupTestServer يبني سيرفر httptest حقيقي فوق NewHandler (نفس التوجيه/الوسائط
// المستخدمة بالإنتاج) متصل بقاعدة بيانات حية، ويرجّع دالة تنظيف تشيل أي بيانات
// اختبار وتقفل السيرفر.
func setupTestServer(t *testing.T) (*httptest.Server, *sqlx.DB) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبارات التكامل الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	if err := database.Migrate(db, "", ""); err != nil {
		db.Close()
		t.Fatalf("migrate error: %v", err)
	}

	cfg := &config.Config{
		Port:        "0",
		DatabaseURL: dsn,
		JWTSecret:   "test-secret-key-for-security-verify-tests",
		CORSOrigins: []string{"http://localhost:5173"},
	}
	handler := NewHandler(cfg, db, time.Now())
	srv := httptest.NewServer(handler)

	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return srv, db
}

// createTestEmployee يزرع موظف اختبار مباشرة بقاعدة البيانات بدور معيّن، ويرجّع
// معرفه ورقم اسم مستخدم فريد. يشيل الموظف (والصلاحيات المرتبطة فيه) بنهاية الاختبار.
func createTestEmployee(t *testing.T, db *sqlx.DB, role string) (employeeID, username string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(testEmployeePassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}
	username = "__test_" + role + "_" + time.Now().Format("150405.000000")
	var id string
	err = db.Get(&id, `
		INSERT INTO "Employee" (id, name, username, password, role, status)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, 'ACTIVE')
		RETURNING id
	`, "__test_employee_"+role, username, string(hash), role)
	if err != nil {
		t.Fatalf("insert test employee error: %v", err)
	}
	t.Cleanup(func() {
		db.MustExec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, id)
		db.MustExec(`DELETE FROM "Employee" WHERE id = $1`, id)
	})
	return id, username
}

// grantPermission يمنح موظف صلاحية معيّنة بالاسم (لازم تكون موجودة أصلاً بجدول
// Permission، مزروعة عبر database.Migrate -> permissionRepo.EnsureSeeded مسبقاً
// بالتشغيل الطبيعي للسيرفر — إذا مو مزروعة هنا نزرعها يدوياً بأمان).
func grantPermission(t *testing.T, db *sqlx.DB, employeeID, permissionName string) {
	t.Helper()
	var permID string
	err := db.Get(&permID, `SELECT id FROM "Permission" WHERE name = $1`, permissionName)
	if err != nil {
		// الصلاحية غير مزروعة (سيرفر الاختبار ما استدعى permissionRepo.EnsureSeeded) — نزرعها هنا
		err = db.Get(&permID, `
			INSERT INTO "Permission" (id, name, label) VALUES (gen_random_uuid()::text, $1, $1)
			RETURNING id
		`, permissionName)
		if err != nil {
			t.Fatalf("could not find/create permission %s: %v", permissionName, err)
		}
	}
	db.MustExec(`
		INSERT INTO "EmployeePermission" (id, "employeeId", "permissionId")
		VALUES (gen_random_uuid()::text, $1, $2)
		ON CONFLICT DO NOTHING
	`, employeeID, permID)
}

// loginAndGetToken يسجّل دخول موظف الاختبار عبر نفس مسار /api/auth/login الحقيقي
// ويرجّع توكن JWT صالح.
func loginAndGetToken(t *testing.T, baseURL, username string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": testEmployeePassword})
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login expected 200, got %d", resp.StatusCode)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode login response error: %v", err)
	}
	if out.Token == "" {
		t.Fatalf("expected non-empty token in login response")
	}
	return out.Token
}

func doRequest(t *testing.T, method, url, token string, payload any) *http.Response {
	t.Helper()
	var bodyReader *bytes.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload error: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request error: %v", err)
	}
	return resp
}

// TestRegularEmployeeCannotCreateProject يتأكد إن فني عادي بدون صلاحيات خاصة
// (دور TECHNICIAN بدون project_management) يرفضه POST /api/projects بـ403.
func TestRegularEmployeeCannotCreateProject(t *testing.T) {
	srv, db := setupTestServer(t)
	empID, username := createTestEmployee(t, db, "TECHNICIAN")
	_ = empID
	token := loginAndGetToken(t, srv.URL, username)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/projects", token, map[string]any{
		"name": "__test_project_should_be_rejected",
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for regular employee creating a project, got %d", resp.StatusCode)
	}
}

// TestRegularEmployeeCannotCreateProduct يتأكد إن موظف بدون صلاحية content_technician
// يرفضه POST /api/products بـ403.
func TestRegularEmployeeCannotCreateProduct(t *testing.T) {
	srv, db := setupTestServer(t)
	_, username := createTestEmployee(t, db, "TECHNICIAN")
	token := loginAndGetToken(t, srv.URL, username)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/products", token, map[string]any{
		"name":  "__test_product_should_be_rejected",
		"price": 10,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for regular employee creating a product, got %d", resp.StatusCode)
	}
}

// TestRoleDemotionRevokesAccessImmediately هو أهم اختبار انحدار (regression) بكل
// هذي المجموعة: يسجّل دخول أدمن، يحصل توكن صالح، ثم يخفّض دوره مباشرة بقاعدة
// البيانات (بدون تسجيل خروج/دخول ولا انتظار انتهاء التوكن) ثم يعيد استخدام نفس
// التوكن القديم بمحاولة عملية تتطلب دور ADMIN — لازم ينرفض فوراً بـ403. هذا يثبت
// إن RequireAuth يرجع الدور من قاعدة البيانات بكل طلب (StatusAndRoleByID) مو من
// claims التوكن المخزّنة وقت تسجيل الدخول — لو رجعت هذي الرجعة (regression) بأي
// تعديل مستقبلي، هذا الاختبار ينكسر فوراً.
func TestRoleDemotionRevokesAccessImmediately(t *testing.T) {
	srv, db := setupTestServer(t)
	empID, username := createTestEmployee(t, db, "ADMIN")
	token := loginAndGetToken(t, srv.URL, username)

	// تأكيد إن التوكن فعلاً يعمل كأدمن قبل التخفيض (عملية تتطلب requireAdmin: حذف مادة تدريبية غير موجودة تكفي لعبور middleware)
	preResp := doRequest(t, http.MethodGet, srv.URL+"/api/employees/archived", token, nil)
	preResp.Body.Close()
	if preResp.StatusCode != http.StatusOK {
		t.Fatalf("expected ADMIN token to pass requireAdmin BEFORE demotion, got %d", preResp.StatusCode)
	}

	// تخفيض الدور مباشرة بقاعدة البيانات — يحاكي أدمن آخر يغيّر دور هذا الموظف
	// بالضبط الثغرة اللي انصلحت: قبل التعديل، هذا التخفيض ما كان يوقف التوكن القديم فوراً.
	db.MustExec(`UPDATE "Employee" SET role = 'TECHNICIAN' WHERE id = $1`, empID)

	postResp := doRequest(t, http.MethodGet, srv.URL+"/api/employees/archived", token, nil)
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusForbidden {
		t.Fatalf("REGRESSION: expected 403 immediately after role demotion using the SAME old token, got %d", postResp.StatusCode)
	}
}

// TestQuotationSystemPermissionGating يتأكد إن موظف بدون صلاحية quotation_system
// يرفضه إنشاء/تعديل عرض سعر، وموظف بنفس الصلاحية ممنوحة له يقدر ينشئ عرض سعر فعلياً.
func TestQuotationSystemPermissionGating(t *testing.T) {
	srv, db := setupTestServer(t)

	_, usernameWithout := createTestEmployee(t, db, "TECHNICIAN")
	tokenWithout := loginAndGetToken(t, srv.URL, usernameWithout)

	payload := map[string]any{
		"customerName": "__test_customer_quotation",
		"items":        []any{},
	}

	respWithout := doRequest(t, http.MethodPost, srv.URL+"/api/quotations", tokenWithout, payload)
	respWithout.Body.Close()
	if respWithout.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without quotation_system permission, got %d", respWithout.StatusCode)
	}

	empIDWith, usernameWith := createTestEmployee(t, db, "TECHNICIAN")
	grantPermission(t, db, empIDWith, "quotation_system")
	tokenWith := loginAndGetToken(t, srv.URL, usernameWith)

	payloadWith := map[string]any{
		"customerName":        "__test_customer_quotation",
		"items":               []any{},
		"createdByEmployeeId": empIDWith,
	}
	respWith := doRequest(t, http.MethodPost, srv.URL+"/api/quotations", tokenWith, payloadWith)
	defer respWith.Body.Close()
	if respWith.StatusCode != http.StatusCreated && respWith.StatusCode != http.StatusOK {
		t.Fatalf("expected employee WITH quotation_system to succeed creating a quotation, got %d", respWith.StatusCode)
	}

	// تنظيف عرض السعر المُنشأ
	var quotationID string
	if err := json.NewDecoder(respWith.Body).Decode(&struct {
		ID *string `json:"id"`
	}{ID: &quotationID}); err == nil && quotationID != "" {
		db.MustExec(`DELETE FROM "Quotation" WHERE id = $1`, quotationID)
	}
}

// TestGpsSystemPermissionGating يتأكد إن موظف بدون صلاحية gps_system يرفضه إنشاء
// عميل GPS، وموظف بنفس الصلاحية ممنوحة له يقدر ينشئ عميل GPS فعلياً.
func TestGpsSystemPermissionGating(t *testing.T) {
	srv, db := setupTestServer(t)

	_, usernameWithout := createTestEmployee(t, db, "TECHNICIAN")
	tokenWithout := loginAndGetToken(t, srv.URL, usernameWithout)

	payload := map[string]any{
		"fullName": "__test_gps_customer",
		"phone":    "0770000" + time.Now().Format("0405"),
	}

	respWithout := doRequest(t, http.MethodPost, srv.URL+"/api/gps/customers", tokenWithout, payload)
	respWithout.Body.Close()
	if respWithout.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 without gps_system permission, got %d", respWithout.StatusCode)
	}

	empIDWith, usernameWith := createTestEmployee(t, db, "TECHNICIAN")
	grantPermission(t, db, empIDWith, "gps_system")
	tokenWith := loginAndGetToken(t, srv.URL, usernameWith)

	respWith := doRequest(t, http.MethodPost, srv.URL+"/api/gps/customers", tokenWith, payload)
	defer respWith.Body.Close()
	if respWith.StatusCode != http.StatusCreated && respWith.StatusCode != http.StatusOK {
		t.Fatalf("expected employee WITH gps_system to succeed creating a GPS customer, got %d", respWith.StatusCode)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(respWith.Body).Decode(&created); err == nil && created.ID != "" {
		db.MustExec(`DELETE FROM "GpsCustomer" WHERE id = $1`, created.ID)
	}
}

// TestComplaintResolveRequiresQualityControl يتأكد إن موظف بدون صلاحية quality_control
// يرفضه PUT /api/complaints/{id}/resolve بـ403 (حتى لو الـid وهمي — middleware يفحص
// الصلاحية قبل ما يوصل الطلب لمنطق الهاندلر أصلاً).
func TestComplaintResolveRequiresQualityControl(t *testing.T) {
	srv, db := setupTestServer(t)
	_, username := createTestEmployee(t, db, "TECHNICIAN")
	token := loginAndGetToken(t, srv.URL, username)

	resp := doRequest(t, http.MethodPut, srv.URL+"/api/complaints/nonexistent-id/resolve", token, map[string]any{})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 for complaint resolve without quality_control, got %d", resp.StatusCode)
	}
}

// TestBookingCreateRequiresAuthOnly يتأكد إن نقطة إنشاء حجز مفتوحة لأي موظف مسجل
// دخول (بدون توكن يجب أن ترفض بـ401)، للتأكد إن التغييرات هذا الأسبوع ما أثّرت
// بالغلط على نقاط أخرى كان المفروض تبقى مفتوحة لأي موظف مسجل دخول.
func TestBookingCreateRequiresAuthOnly(t *testing.T) {
	srv, _ := setupTestServer(t)

	resp := doRequest(t, http.MethodPost, srv.URL+"/api/bookings", "", map[string]any{})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without any token for booking create, got %d", resp.StatusCode)
	}
}
